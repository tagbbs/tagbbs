package rkv

import "encoding/json"

// S is Structured rkv. It's a layer on top of Interface for data structures.
// Currently the only
type S struct {
	Interface
}

func (s S) Read(key string, rev *int64, v interface{}) (err error) {
	var p Value
	p, err = s.Get(key)
	if err != nil {
		return
	}
	if rev != nil {
		*rev = p.Rev
	}
	if len(p.Content) > 0 {
		err = json.Unmarshal(p.Content, v)
	}
	return
}

func (s S) Write(key string, rev int64, v interface{}) (err error) {
	p := Value{Rev: rev}
	p.Content, err = json.Marshal(v)
	if err != nil {
		return
	}
	return s.Put(key, p)
}

func (s S) ReadModify(key string, v interface{}, mutate func(v interface{}) bool) (err error) {
	var rev int64
begin:
	// read value
	err = s.Read(key, &rev, v)
	if err != nil {
		return
	}

	// mutate value
	if mutate == nil || !mutate(v) {
		return nil
	}

	// put value if modified.
	err = s.Write(key, rev+1, v)
	if err == ErrRevNotMatch {
		goto begin
	} else {
		return err
	}
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
func (s S) SetDelete(key string, values ...string) error {
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
