package rkv

import "time"

type MemStore map[string]Value

func (b MemStore) Get(key string) (Value, error) {
	p := b[key]
	content := make([]byte, len(p.Content))
	copy(content, p.Content)
	p.Content = content
	return p, nil
}

func (m MemStore) Put(key string, np Value) error {
	post := m[key]
	if post.Rev+1 != np.Rev {
		return ErrRevNotMatch
	}
	np.Timestamp = time.Now()

	m[key] = np
	return nil
}
