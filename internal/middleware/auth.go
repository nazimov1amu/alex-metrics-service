package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type BufferedResponseWriter struct {
	w      http.ResponseWriter
	body   bytes.Buffer
	status int
}

func (rw *BufferedResponseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *BufferedResponseWriter) WriteHeader(status int) {
	rw.status = status
}

func (rw *BufferedResponseWriter) Write(data []byte) (int, error) {
	return rw.body.Write(data)
}

func (rw *BufferedResponseWriter) Commit(key string) error {
	mac := hmac.New(sha256.New, []byte(key))

	_, _ = mac.Write(rw.body.Bytes())

	signature := hex.EncodeToString(mac.Sum(nil))

	rw.w.Header().Set("HashSHA256", signature)

	rw.w.WriteHeader(rw.status)

	_, err := rw.w.Write(rw.body.Bytes())

	return err
}

func NewBufferedResponseWriter(w http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{w: w}
}

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
