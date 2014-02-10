package tagbbs

import "strings"

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
	m[key] = np
	return nil
}

func (m MemStore) Enumerate(prefix string, pp func(Post)) error {
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			pp(v)
		}
	}
	return nil
}
