package tagbbs

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"

	"time"
)

var (
	ErrAccessDenied = errors.New("Access Denied")
	ErrUserExisted  = errors.New("User Already Existed")
)

var RootUser = "sysop"

type BBS struct {
	store Storage
}

func NewBBS(store Storage) *BBS {
	b := &BBS{store: store}
	b.init()
	return b
}

func (b *BBS) Get(key string, user string) (Post, error) {
	p, err := b.store.Get(key)
	if err != nil {
		return Post{}, err
	}
	if !b.allow(key, p, user, false) {
		return Post{}, ErrAccessDenied
	}
	return p, err
}

func (b *BBS) Put(key string, post Post, user string) error {
	post.Timestamp = time.Now()
	oldpost, err := b.store.Get(key)
	if err != nil {
		return err
	}
	if !b.allow(key, oldpost, user, true) {
		return ErrAccessDenied
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

func (b *BBS) allow(key string, post Post, user string, write bool) bool {
	if user == RootUser {
		return true
	}

	users := []string{}
	check(b.meta("users", &users, nil))
	if i := sort.StringSlice(users).Search(user); i == len(users) || users[i] != user {
		return false
	}

	// deal with post
	if strings.HasPrefix(key, "post:") {
		if !write {
			return true
		} else {
			// new post
			if post.Rev == 0 {
				return true
			}
			fm := post.FrontMatter()
			if fm != nil {
				for _, a := range fm.Authors {
					if a == user {
						return true
					}
				}
			}
		}
	}

	return false
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

func (b *BBS) NewUser(user string) error {
	var (
		users []string
		err   error
	)

	if err2 := b.meta("users", &users, func(v interface{}) {
		if i := sort.StringSlice(users).Search(user); i < len(users) && users[i] == user {
			err = ErrUserExisted
		} else {
			users = append(users[:i], append([]string{user}, users[i:]...)...)
		}
	}); err2 != nil {
		return err2
	}

	return err
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
