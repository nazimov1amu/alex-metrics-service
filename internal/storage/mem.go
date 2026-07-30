package storage

import "errors"


type MemStorage[T any] struct {
	storage map[string]T
}

func NewMemStorage[T any]() *MemStorage[T] {
	return &MemStorage[T]{
		storage: make(map[string]T),
	}
}

func (s *MemStorage[T]) Get(key string) (T, error) {
	value, ok := s.storage[key]
	if !ok {
		return value, errors.New("value not found")
	}
	return value, nil
}

func (s *MemStorage[T]) Update(key string, value T) error {
	s.storage[key] = value
	return nil
}

func (s *MemStorage[T]) GetBulk() ([]T, error) {
	values := make([]T, 0, len(s.storage))
	for _, value := range s.storage {
		values = append(values, value)
	}
	return values, nil
}