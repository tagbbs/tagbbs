package rkv

import (
	"sync"

	"time"
)

type MemStore struct {
	kvs map[string]Value
	mu  sync.Mutex
}

func NewMemStore() *MemStore {
	return &MemStore{kvs: make(map[string]Value)}
}

func (m *MemStore) Get(key string) (Value, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.kvs[key]
	content := make([]byte, len(p.Content))
	copy(content, p.Content)
	p.Content = content
	return p, nil
}

func (m *MemStore) Put(key string, np Value) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	post := m.kvs[key]
	if post.Rev+1 != np.Rev {
		return ErrRevNotMatch
	}
	np.Timestamp = time.Now()

	m.kvs[key] = np
	return nil
}
