package rkv

type ScopedStore struct {
	Interface
	Prefix string
}

func (s ScopedStore) Get(key string) (Value, error) {
	return s.Interface.Get(s.Prefix + key)
}

func (s ScopedStore) Put(key string, v Value) error {
	return s.Interface.Put(s.Prefix+key, v)
}
