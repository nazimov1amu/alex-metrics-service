package repository

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

type FileAuditRepository struct {
	path string
	mu   sync.Mutex
}

func NewFileAuditRepository(path string) *FileAuditRepository {
	return &FileAuditRepository{path: path}
}

func (r *FileAuditRepository) SaveAudit(audit *model.Audit) error {
	data, err := json.Marshal(audit)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}
