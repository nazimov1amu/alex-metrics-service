// Package model описывает метрики и события аудита, которыми обмениваются слои сервиса.
package model

const (
	// Counter — тип метрики, значение которой суммируется при каждом обновлении.
	Counter = "counter"
	// Gauge — тип метрики, значение которой заменяется при каждом обновлении.
	Gauge = "gauge"
)

// Metrics — одна gauge- или counter-метрика в текстовом и JSON API.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
