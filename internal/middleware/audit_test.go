package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/handler"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newAuditRouter(t *testing.T, channels []chan *model.Audit) http.Handler {
	t.Helper()
	repo := repository.NewMemMetricsRepository()
	svc := service.NewMetricsService(repo, &config.ServerConfig{})
	h := handler.NewMetricsHandler(svc, zap.NewNop().Sugar())
	metrics := handler.MetricsRouter(h, AuditMiddleware(channels))
	return handler.NewGlobalRouter(nil, []handler.Mount{{Pattern: "/", Router: metrics}})
}

func waitAudit(t *testing.T, ch chan *model.Audit) *model.Audit {
	t.Helper()
	select {
	case event := <-ch:
		return event
	case <-time.After(time.Second):
		t.Fatal("audit event was not delivered")
		return nil
	}
}

func TestAuditMiddleware_SendsEventToEveryChannel(t *testing.T) {
	fileCh := make(chan *model.Audit, 1)
	httpCh := make(chan *model.Audit, 1)
	router := newAuditRouter(t, []chan *model.Audit{fileCh, httpCh})

	body := []byte(`[{"id":"Alloc","type":"gauge","value":1},{"id":"Frees","type":"gauge","value":2}]`)
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	req.Header.Set("X-Real-IP", "192.168.0.42")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	for _, ch := range []chan *model.Audit{fileCh, httpCh} {
		event := waitAudit(t, ch)
		assert.Equal(t, []string{"Alloc", "Frees"}, event.Metrics)
		assert.Equal(t, "192.168.0.42", event.IPAddress)
		assert.NotZero(t, event.Timestamp)
	}
}

func TestAuditMiddleware_UpdatePath(t *testing.T) {
	ch := make(chan *model.Audit, 1)
	router := newAuditRouter(t, []chan *model.Audit{ch})

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/1.5", nil)
	req.Header.Set("X-Real-IP", "10.0.0.8")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	event := waitAudit(t, ch)
	assert.Equal(t, []string{"Alloc"}, event.Metrics)
	assert.Equal(t, "10.0.0.8", event.IPAddress)
}

func TestAuditMiddleware_SkipsFailedUpdate(t *testing.T) {
	ch := make(chan *model.Audit, 1)
	router := newAuditRouter(t, []chan *model.Audit{ch})

	req := httptest.NewRequest(http.MethodPost, "/update/unknown/Alloc/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	select {
	case event := <-ch:
		t.Fatalf("unexpected audit event: %+v", event)
	default:
	}
}
