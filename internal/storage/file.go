package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FileStorage[T any] struct {
	mem       *MemStorage[T]
	path      string
	syncWrite bool
}

func NewFileStorage[T any](path string, restore bool, syncWrite bool) (*FileStorage[T], error) {
	fileStorage := &FileStorage[T]{mem: NewMemStorage[T](), path: path, syncWrite: syncWrite}
	if restore {
		if err := fileStorage.load(); err != nil {
			return nil, err
		}
	}
	return fileStorage, nil
}

func (s *FileStorage[T]) Get(key string) (T, error) {
	return s.mem.Get(key)
}

func (s *FileStorage[T]) GetBulk() ([]T, error) {
	return s.mem.GetBulk()
}

func (s *FileStorage[T]) Update(key string, value T) error {
	if err := s.mem.Update(key, value); err != nil {
		return err
	}
	if s.syncWrite {
		return s.Store()
	}
	return nil
}

func (s *FileStorage[T]) Store() error {
	dir := filepath.Dir(s.path)

	tmp, err := os.CreateTemp(dir, "metrics-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	s.mem.mutex.Lock()
	data := make(map[string]T, len(s.mem.storage))
	for k, v := range s.mem.storage {
		data[k] = v
	}
	s.mem.mutex.Unlock()

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

	return os.Rename(tmpName, s.path)
}

func (s *FileStorage[T]) load() error {
	file, err := os.Open(s.path)
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

	var data map[string]T
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	s.mem.mutex.Lock()
	defer s.mem.mutex.Unlock()
	if data == nil {
		s.mem.storage = make(map[string]T)
	} else {
		s.mem.storage = data
	}
	return nil
}
