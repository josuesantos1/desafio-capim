// Package storage provides a generic, concurrency-safe in-memory key-value store.
package storage

import (
	"errors"
	"sync"
)

var (
	ErrNotFound      = errors.New("storage: not found")
	ErrAlreadyExists = errors.New("storage: already exists")
)

type Store[V any] struct {
	mu   sync.RWMutex
	data map[string]V
}

func New[V any]() *Store[V] {
	return &Store[V]{data: make(map[string]V)}
}

func (s *Store[V]) Insert(id string, v V) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[id]; ok {
		return ErrAlreadyExists
	}
	s.data[id] = v
	return nil
}

func (s *Store[V]) Read(id string) (V, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[id]
	if !ok {
		var zero V
		return zero, ErrNotFound
	}
	return v, nil
}

func (s *Store[V]) Update(id string, v V) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	s.data[id] = v
	return nil
}

func (s *Store[V]) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	delete(s.data, id)
	return nil
}

func (s *Store[V]) All() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]V, 0, len(s.data))
	for _, v := range s.data {
		out = append(out, v)
	}
	return out
}
