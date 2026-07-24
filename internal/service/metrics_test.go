package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/Alexunder2003/alex-metrics-service/internal/storage"
)

func ptr[T any](v T) *T {
	return &v
}

func TestMetricsService_Update(t *testing.T) {
	tests := []struct {
		name    string
		input   model.MetricsInput
		want    model.Metrics
		wantErr error
	}{
		{
			name:  "update counter",
			input: model.MetricsInput{Name: "test_counter", MType: model.Counter, RawValue: "100"},
			want:  model.Metrics{ID: "test_counter", MType: model.Counter, Delta: ptr(int64(100))},
		},
		{
			name:  "update gauge",
			input: model.MetricsInput{Name: "test_gauge", MType: model.Gauge, RawValue: "100.5"},
			want:  model.Metrics{ID: "test_gauge", MType: model.Gauge, Value: ptr(float64(100.5))},
		},
		{
			name:    "update counter with wrong value",
			input:   model.MetricsInput{Name: "test_counter", MType: model.Counter, RawValue: "asdf"},
			wantErr: ErrInvalidCounterValue,
		},
		{
			name:    "update gauge with negative value",
			input:   model.MetricsInput{Name: "test_gauge", MType: model.Gauge, RawValue: "asdf"},
			wantErr: ErrInvalidGaugeValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage[model.Metrics]()
			svc := NewMetricsService(store)

			got, err := svc.Update(tt.input)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.Error(t, err)
			}

			stored, err := svc.Get(tt.input.Name)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, stored)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMetricsService_Get(t *testing.T) {
	tests := []struct {
		name    string
		input   model.MetricsInput
		want    model.Metrics
		wantErr error
	}{
		{
			name:  "get counter",
			input: model.MetricsInput{Name: "test_counter", MType: model.Counter, RawValue: "100"},
			want:  model.Metrics{ID: "test_counter", MType: model.Counter, Delta: ptr(int64(100))},
		},
		{
			name:  "get gauge",
			input: model.MetricsInput{Name: "test_gauge", MType: model.Gauge, RawValue: "100.5"},
			want:  model.Metrics{ID: "test_gauge", MType: model.Gauge, Value: ptr(float64(100.5))},
		},
		{
			name:    "get counter not found",
			wantErr: ErrMetricNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage[model.Metrics]()
			svc := NewMetricsService(store)

			if tt.input != (model.MetricsInput{}) {
				_, err := svc.Update(tt.input)
				assert.NoError(t, err)
			}

			stored, err := svc.Get(tt.input.Name)
			if tt.wantErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, stored)
		})
	}
}

func TestMetricsService_GetBulk(t *testing.T) {
	tests := []struct {
		name   string
		inputs []model.MetricsInput
		want   []model.Metrics
	}{
		{
			name:   "empty storage",
			inputs: nil,
			want:   []model.Metrics{},
		},
		{
			name: "single counter",
			inputs: []model.MetricsInput{
				{Name: "test_counter", MType: model.Counter, RawValue: "100"},
			},
			want: []model.Metrics{
				{ID: "test_counter", MType: model.Counter, Delta: ptr(int64(100))},
			},
		},
		{
			name: "single gauge",
			inputs: []model.MetricsInput{
				{Name: "test_gauge", MType: model.Gauge, RawValue: "100.5"},
			},
			want: []model.Metrics{
				{ID: "test_gauge", MType: model.Gauge, Value: ptr(float64(100.5))},
			},
		},
		{
			name: "multiple metrics",
			inputs: []model.MetricsInput{
				{Name: "test_counter", MType: model.Counter, RawValue: "100"},
				{Name: "test_gauge", MType: model.Gauge, RawValue: "100.5"},
			},
			want: []model.Metrics{
				{ID: "test_counter", MType: model.Counter, Delta: ptr(int64(100))},
				{ID: "test_gauge", MType: model.Gauge, Value: ptr(float64(100.5))},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage[model.Metrics]()
			svc := NewMetricsService(store)

			for _, input := range tt.inputs {
				_, err := svc.Update(input)
				assert.NoError(t, err)
			}

			got, err := svc.GetBulk()
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
