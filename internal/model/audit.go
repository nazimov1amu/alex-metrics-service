package model

// Audit — событие об успешном изменении метрик: время, имена и IP клиента.
type Audit struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
