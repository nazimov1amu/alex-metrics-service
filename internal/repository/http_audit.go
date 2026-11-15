package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

// HTTPAuditRepository отправляет события аудита POST-запросом на внешний URL.
type HTTPAuditRepository struct {
	url    string
	client *http.Client
}

// NewHTTPAuditRepository создаёт клиент аудита с таймаутом 5 секунд.
func NewHTTPAuditRepository(url string) *HTTPAuditRepository {
	return &HTTPAuditRepository{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// SaveAudit отправляет событие аудита JSON-телом и считает успешным только ответ 200.
func (r *HTTPAuditRepository) SaveAudit(audit *model.Audit) error {
	body, err := json.Marshal(audit)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, r.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("audit post failed: %s", resp.Status)
	}
	return nil
}
