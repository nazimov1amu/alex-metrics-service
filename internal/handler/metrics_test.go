package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
)

func ptr[T any](v T) *T { return &v }

func newTestRouter() http.Handler {
	repo := repository.NewMemMetricsRepository()
	svc := service.NewMetricsService(repo, &config.ServerConfig{})
	return MetricsRouter(NewMetricsHandler(svc, zap.NewNop().Sugar()))
}

func TestMetricsHandler_UpdatePath(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "update counter",
			url:        "/update/counter/test_counter/100",
			wantStatus: http.StatusOK,
			wantBody:   "100",
		},
		{
			name:       "update gauge",
			url:        "/update/gauge/test_gauge/100.5",
			wantStatus: http.StatusOK,
			wantBody:   "100.500000",
		},
		{
			name:       "invalid type",
			url:        "/update/unknown/test/1",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid metric type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestMetricsHandler_ValuePath(t *testing.T) {
	tests := []struct {
		name       string
		seedURL    string
		url        string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "get counter",
			seedURL:    "/update/counter/test_counter/100",
			url:        "/value/counter/test_counter",
			wantStatus: http.StatusOK,
			wantBody:   "100",
		},
		{
			name:       "get gauge",
			seedURL:    "/update/gauge/test_gauge/100.5",
			url:        "/value/gauge/test_gauge",
			wantStatus: http.StatusOK,
			wantBody:   "100.5",
		},
		{
			name:       "metric not found",
			url:        "/value/counter/missing",
			wantStatus: http.StatusNotFound,
			wantBody:   "metric not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter()
			if tt.seedURL != "" {
				seedRec := httptest.NewRecorder()
				seedReq := httptest.NewRequest(http.MethodPost, tt.seedURL, nil)
				r.ServeHTTP(seedRec, seedReq)
				require.Equal(t, http.StatusOK, seedRec.Code)
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestMetricsHandler_UpdateJSON(t *testing.T) {
	r := newTestRouter()

	body, err := json.Marshal(model.Metrics{
		ID:    "Alloc",
		MType: model.Gauge,
		Value: ptr(123.45),
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var got model.Metrics
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "Alloc", got.ID)
	assert.Equal(t, model.Gauge, got.MType)
	require.NotNil(t, got.Value)
	assert.InDelta(t, 123.45, *got.Value, 0.0001)
}

func TestMetricsHandler_ValueJSON(t *testing.T) {
	r := newTestRouter()

	updateBody, err := json.Marshal(model.Metrics{
		ID:    "PollCount",
		MType: model.Counter,
		Delta: ptr(int64(7)),
	})
	require.NoError(t, err)

	updateRec := httptest.NewRecorder()
	updateReq := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(updateBody))
	r.ServeHTTP(updateRec, updateReq)
	require.Equal(t, http.StatusOK, updateRec.Code)

	valueBody, err := json.Marshal(model.Metrics{
		ID:    "PollCount",
		MType: model.Counter,
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(valueBody))
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got model.Metrics
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.NotNil(t, got.Delta)
	assert.Equal(t, int64(7), *got.Delta)
}
