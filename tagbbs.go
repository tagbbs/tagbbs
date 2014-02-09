package tagbbs

import (
	"encoding/json"
	"strconv"
	"strings"

	"time"
)

type BBS struct {
	store Storage
}

func NewBBS(store Storage) *BBS {
	b := &BBS{store: store}
	b.init()
	return b
}

func (b *BBS) Get(key string) (Post, error) {
	// TODO: Permission checking
	return b.store.Get(key)
}

func (b *BBS) Put(key string, post Post) error {
	// TODO: Permission checking
	post.Timestamp = time.Now()
	oldpost, err := b.store.Get(key)
	if err != nil {
		return err
	}
	if err := b.store.Put(key, post); err != nil {
		return err
	}
	if strings.HasPrefix(key, "post:") {
		b.indexRemove(key, oldpost)
		b.indexAdd(key, post)
	}
	return nil
}

func (b *BBS) meta(key string, value interface{}, mutate func(v interface{})) error {
	return b.modify("bbs:"+key, value, mutate)
}

func (b *BBS) modify(key string, value interface{}, mutate func(v interface{})) error {
begin:
	// read value
	p, err := b.store.Get(key)
	if err != nil {
		return err
	}
	if len(p.Content) > 0 {
		if err := json.Unmarshal(p.Content, value); err != nil {
			return err
		}
	}

	// mutate value
	if mutate == nil {
		return nil
	}
	mutate(value)

	// write value
	if p.Content, err = json.Marshal(value); err != nil {
		return err
	}
	p.Rev++
	p.Timestamp = time.Now()
	err = b.store.Put(key, p)
	if err == ErrRevNotMatch {
		goto begin
	} else {
		return err
	}
}

func (b *BBS) NewPostKey() string {
	var (
		nextid int64
		key    string
	)
	check(b.meta("nextid", &nextid, func(v interface{}) {
		key = "post:" + strconv.FormatInt(nextid, 16)
		nextid++
	}))
	return key
}

func (b *BBS) init() {
	var (
		nextid int64
		name   string
	)
	check(b.meta("nextid", &nextid, func(v interface{}) {}))
	check(b.meta("name", &name, func(v interface{}) {
		if len(name) == 0 {
			name = "newbbs"
		}
	}))
}

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
