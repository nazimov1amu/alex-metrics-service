package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
)

func ptr[T any](v T) *T {
	return &v
}

func TestMetricsService_Update(t *testing.T) {
	tests := []struct {
		name    string
		metric  *model.Metrics
		want    model.Metrics
		wantErr error
	}{
		{
			name:   "update counter",
			metric: &model.Metrics{ID: "test_counter", MType: model.Counter, Delta: ptr(int64(100))},
			want:   model.Metrics{ID: "test_counter", MType: model.Counter, Delta: ptr(int64(100))},
		},
		{
			name:   "update gauge",
			metric: &model.Metrics{ID: "test_gauge", MType: model.Gauge, Value: ptr(float64(100.5))},
			want:   model.Metrics{ID: "test_gauge", MType: model.Gauge, Value: ptr(float64(100.5))},
		},
		{
			name:    "update with invalid type",
			metric:  &model.Metrics{ID: "test", MType: "unknown", Value: ptr(float64(1))},
			wantErr: ErrInvalidMetricType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewMemMetricsRepository()
			svc := NewMetricsService(repo, &config.ServerConfig{})

			err := svc.Update(context.Background(), tt.metric)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, *tt.metric)

			got, err := svc.Get(context.Background(), tt.metric.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetricsService_Update_CounterAccumulates(t *testing.T) {
	repo := repository.NewMemMetricsRepository()
	svc := NewMetricsService(repo, &config.ServerConfig{})

	first := &model.Metrics{ID: "poll", MType: model.Counter, Delta: ptr(int64(10))}
	require.NoError(t, svc.Update(context.Background(), first))
	assert.Equal(t, int64(10), *first.Delta)

	second := &model.Metrics{ID: "poll", MType: model.Counter, Delta: ptr(int64(5))}
	require.NoError(t, svc.Update(context.Background(), second))
	assert.Equal(t, int64(15), *second.Delta)

	got, err := svc.Get(context.Background(), "poll")
	require.NoError(t, err)
	assert.Equal(t, int64(15), *got.Delta)
}

func TestMetricsService_Get(t *testing.T) {
	tests := []struct {
		name    string
		seed    *model.Metrics
		query   model.Metrics
		want    model.Metrics
		wantErr error
	}{
		{
			name:  "get counter",
			seed:  &model.Metrics{ID: "test_counter", MType: model.Counter, Delta: ptr(int64(100))},
			query: model.Metrics{ID: "test_counter"},
			want:  model.Metrics{ID: "test_counter", MType: model.Counter, Delta: ptr(int64(100))},
		},
		{
			name:  "get gauge",
			seed:  &model.Metrics{ID: "test_gauge", MType: model.Gauge, Value: ptr(float64(100.5))},
			query: model.Metrics{ID: "test_gauge"},
			want:  model.Metrics{ID: "test_gauge", MType: model.Gauge, Value: ptr(float64(100.5))},
		},
		{
			name:    "get not found",
			query:   model.Metrics{ID: "missing"},
			wantErr: ErrMetricNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := repository.NewFileMetricsRepository(t.TempDir()+"/metrics.json", false, false)
			require.NoError(t, err)
			svc := NewMetricsService(repo, &config.ServerConfig{})

			if tt.seed != nil {
				require.NoError(t, svc.Update(context.Background(), tt.seed))
			}

			got, err := svc.Get(context.Background(), tt.query.ID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
