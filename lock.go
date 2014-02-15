package tagbbs

import "errors"

var (
	ErrNotLocked     = errors.New("Not Locked")
	ErrAlreadyLocked = errors.New("Already Locked")
)

func (b *BBS) lock(key string) error {
	key = "_lock:" + key
begin:
	p, err := b.store.Get(key)
	if err != nil {
		return err
	}
	if len(p.Content) != 0 {
		return ErrAlreadyLocked
	}
	p.Rev++
	p.Content = []byte("!")
	err = b.store.Put(key, p)
	if err == ErrRevNotMatch {
		goto begin
	} else {
		return err
	}
}

func (b *BBS) unlock(key string) error {
	key = "_lock:" + key
begin:
	p, err := b.store.Get(key)
	if len(p.Content) == 0 {
		return ErrNotLocked
	}
	p.Rev++
	p.Content = nil
	err = b.store.Put(key, p)
	if err == ErrRevNotMatch {
		goto begin
	} else {
		return err
	}
}
