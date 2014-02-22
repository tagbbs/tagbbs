package rkv

type ScopedStore struct {
	Storage
	Prefix string
}

func (s ScopedStore) Get(key string) (Value, error) {
	return s.Storage.Get(s.Prefix + key)
}

func (s ScopedStore) Put(key string, v Value) error {
	return s.Storage.Put(s.Prefix+key, v)
}
