package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// BufferedResponseWriter копит статус и тело ответа, чтобы подписать их до отправки клиенту.
type BufferedResponseWriter struct {
	w      http.ResponseWriter
	body   bytes.Buffer
	status int
}

// Header возвращает заголовки исходного ResponseWriter.
func (rw *BufferedResponseWriter) Header() http.Header {
	return rw.w.Header()
}

// WriteHeader запоминает статус и не отправляет его, пока не вызван Commit.
func (rw *BufferedResponseWriter) WriteHeader(status int) {
	rw.status = status
}

// Write дописывает тело во внутренний буфер.
func (rw *BufferedResponseWriter) Write(data []byte) (int, error) {
	return rw.body.Write(data)
}

// Commit подписывает накопленное тело HMAC-SHA256, ставит заголовок HashSHA256 и отправляет ответ.
func (rw *BufferedResponseWriter) Commit(key string) error {
	mac := hmac.New(sha256.New, []byte(key))

	_, _ = mac.Write(rw.body.Bytes())

	signature := hex.EncodeToString(mac.Sum(nil))

	rw.w.Header().Set("HashSHA256", signature)

	rw.w.WriteHeader(rw.status)

	_, err := rw.w.Write(rw.body.Bytes())

	return err
}

// NewBufferedResponseWriter создаёт буфер ответа поверх w.
func NewBufferedResponseWriter(w http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{w: w}
}

// AuthMiddleware проверяет заголовок HashSHA256 запроса и подписывает ответ тем же ключом.
// Пустой key отключает проверку, и запрос проходит дальше без подписи.
func AuthMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			sign := r.Header.Get("HashSHA256")
			if sign == "" {
				http.Error(w, "header is required", http.StatusBadRequest)
				return
			}

			mac := hmac.New(sha256.New, []byte(key))
			mac.Write(body)
			expected := hex.EncodeToString(mac.Sum(nil))

			if !hmac.Equal([]byte(expected), []byte(sign)) {
				http.Error(w, "invalid hash", http.StatusBadRequest)
				return
			}

			rw := NewBufferedResponseWriter(w)

			next.ServeHTTP(rw, r)

			err = rw.Commit(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}
