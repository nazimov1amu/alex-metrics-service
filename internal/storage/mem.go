package storage

import (
	"errors"
	"sync"
)


type MemStorage[T any] struct {
	storage map[string]T
	mutex sync.Mutex
}

func NewMemStorage[T any]() *MemStorage[T] {
	return &MemStorage[T]{storage: make(map[string]T), mutex: sync.Mutex{}}
}

func (s *MemStorage[T]) Get(key string) (T, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	value, ok := s.storage[key]
	if !ok {
		return value, errors.New("value not found")
	}
	return value, nil
}

func (s *MemStorage[T]) Update(key string, value T) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.storage[key] = value
	return nil
}

func (s *MemStorage[T]) GetBulk() ([]T, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	values := make([]T, 0, len(s.storage))
	for _, value := range s.storage {
		values = append(values, value)
	}
	return values, nil
}

