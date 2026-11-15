package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

// FileMetricsRepository держит метрики в памяти и сохраняет их в JSON-файл.
type FileMetricsRepository struct {
	mem       *MemMetricsRepository
	path      string
	syncWrite bool
}

// NewFileMetricsRepository создаёт файловое хранилище.
// При restore == true метрики читаются из path. При syncWrite == true каждая запись сразу сбрасывается на диск.
func NewFileMetricsRepository(path string, restore bool, syncWrite bool) (*FileMetricsRepository, error) {
	repo := &FileMetricsRepository{
		mem:       NewMemMetricsRepository(),
		path:      path,
		syncWrite: syncWrite,
	}
	if restore {
		if err := repo.load(); err != nil {
			return nil, err
		}
	}
	return repo, nil
}

// Get возвращает метрику по идентификатору.
func (r *FileMetricsRepository) Get(ctx context.Context, id string) (model.Metrics, error) {
	return r.mem.Get(ctx, id)
}

// Update сохраняет метрику и при синхронном режиме сразу записывает файл.
func (r *FileMetricsRepository) Update(ctx context.Context, metric model.Metrics) error {
	if err := r.mem.Update(ctx, metric); err != nil {
		return err
	}
	if r.syncWrite {
		return r.Store()
	}
	return nil
}

// GetBulk возвращает все метрики из памяти.
func (r *FileMetricsRepository) GetBulk(ctx context.Context) ([]model.Metrics, error) {
	return r.mem.GetBulk(ctx)
}

// Store сериализует текущие метрики в JSON-файл хранилища.
func (r *FileMetricsRepository) Store() error {
	dir := filepath.Dir(r.path)

	tmp, err := os.CreateTemp(dir, "metrics-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	r.mem.mutex.Lock()
	data := make(map[string]model.Metrics, len(r.mem.storage))
	for k, v := range r.mem.storage {
		data[k] = v
	}
	r.mem.mutex.Unlock()

	if err := json.NewEncoder(tmp).Encode(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, r.path)
}

func (r *FileMetricsRepository) load() error {
	file, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return nil
	}

	var data map[string]model.Metrics
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	r.mem.mutex.Lock()
	defer r.mem.mutex.Unlock()
	if data == nil {
		r.mem.storage = make(map[string]model.Metrics)
	} else {
		r.mem.storage = data
	}
	return nil
}

// BulkUpdate сохраняет набор метрик и записывает файл.
func (r *FileMetricsRepository) BulkUpdate(ctx context.Context, metrics []model.Metrics) error {
	if err := r.mem.BulkUpdate(ctx, metrics); err != nil {
		return err
	}
	if err := r.Store(); err != nil {
		return err
	}
	return nil
}
