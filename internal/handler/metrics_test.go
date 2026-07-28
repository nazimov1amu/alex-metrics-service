package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
	"github.com/Alexunder2003/alex-metrics-service/internal/storage"
)

func TestMetricsHandler_Update(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name  string
		url   string
		want  want
	}{
		{
			name: "update counter",
			url:  "/update/counter/test_counter/100",
			want: want{statusCode: http.StatusOK, body: "100"},
		},
		{
			name: "update gauge",
			url:  "/update/gauge/test_gauge/100.5",
			want: want{statusCode: http.StatusOK, body: "100.500000"},
		},
		{
			name: "invalid type",
			url:  "/update/unknown/test/1",
			want: want{statusCode: http.StatusBadRequest, body: "invalid metric type\n"},
		},
	}

	svc := service.NewMetricsService(storage.NewMemStorage[model.Metrics]())
	r := MetricsRouter(New(svc))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.want.statusCode, rec.Code)
			assert.Equal(t, tt.want.body, rec.Body.String())
		})
	}
}

func TestMetricsHandler_Get(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name       string
		seedURL    string
		url        string
		want       want
	}{
		{
			name:    "get counter",
			seedURL: "/update/counter/test_counter/100",
			url:     "/value/counter/test_counter",
			want:    want{statusCode: http.StatusOK, body: "100"},
		},
		{
			name:    "get gauge",
			seedURL: "/update/gauge/test_gauge/100.5",
			url:     "/value/gauge/test_gauge",
			want:    want{statusCode: http.StatusOK, body: "100.5"},
		},
		{
			name: "metric not found",
			url:  "/value/counter/missing",
			want: want{statusCode: http.StatusNotFound, body: "metric not found\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewMetricsService(storage.NewMemStorage[model.Metrics]())
			r := MetricsRouter(New(svc))

			if tt.seedURL != "" {
				seedRec := httptest.NewRecorder()
				seedReq := httptest.NewRequest(http.MethodPost, tt.seedURL, nil)
				r.ServeHTTP(seedRec, seedReq)
				assert.Equal(t, http.StatusOK, seedRec.Code)
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.want.statusCode, rec.Code)
			assert.Equal(t, tt.want.body, rec.Body.String())
		})
	}
}

