package rkv

import "encoding/json"

// S is Structured rkv. It's a layer on top of Interface for data structures.
// Currently the only
type S struct {
	Interface
}

func (s S) ReadModify(key string, v interface{}, mutate func(v interface{}) bool) error {
begin:
	// read value
	p, err := s.Get(key)
	if err != nil {
		return err
	}
	if len(p.Content) > 0 {
		if err := json.Unmarshal(p.Content, v); err != nil {
			return err
		}
	}

	// mutate value
	if mutate == nil || !mutate(v) {
		return nil
	}

	// put value if modified.
	if p.Content, err = json.Marshal(v); err != nil {
		return err
	}
	p.Rev++
	err = s.Put(key, p)
	if err == ErrRevNotMatch {
		goto begin
	} else {
		return err
	}
}

func (s S) Read(key string, v interface{}) error {
	return s.ReadModify(key, v, nil)
}

// Sorted Set of strings.

// SetInsert will insert the given values to the key.
func (s S) SetInsert(key string, values ...string) error {
	var vv SortedString
	return s.ReadModify(key, &vv, func(_ interface{}) (r bool) {
		for _, v := range values {
			r = r || vv.Insert(v)
		}
		return
	})
}

// SetDelete will delete the given values from the key.
func (s *S) SetDelete(key string, values ...string) error {
	var vv SortedString
	return s.ReadModify(key, &vv, func(_ interface{}) (r bool) {
		for _, v := range values {
			r = r || vv.Delete(v)
		}
		return
	})
}

// SetContain return true if all values exists in the set.
func (s S) SetContain(key string, values ...string) (contain bool, err error) {
	var vv SortedString
	contain = true
	err = s.ReadModify(key, &vv, func(_ interface{}) (r bool) {
		for _, v := range values {
			contain = contain && vv.Contain(v)
		}
		return
	})
	return
}

// SetSlice return the slice around the value.
// Suppose the value is at position i, the returned slice will be [i-before:i+after].
// If i == 0, the result will be [len-before:].
func (s S) SetSlice(key, value string, before, after int) (slice []string, err error) {
	var vv SortedString
	err = s.ReadModify(key, &vv, func(_ interface{}) (r bool) {
		slice = vv.Slice(value, before, after)
		return
	})
	return
}
