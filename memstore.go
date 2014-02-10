package tagbbs

import "time"

type MemStore map[string]Post

func (b MemStore) Get(key string) (Post, error) {
	p := b[key]
	return p, nil
}

func (m MemStore) Put(key string, np Post) error {
	post := m[key]
	if post.Rev+1 != np.Rev {
		return ErrRevNotMatch
	}
	np.Timestamp = time.Now()

	m[key] = np
	return nil
}
