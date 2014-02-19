package tagbbs

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"

	"time"
)

var (
	ErrAccessDenied = errors.New("Access Denied")
	ErrUserExists   = errors.New("User Exists")
)

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
		b.indexReplace(key, oldpost, post)
	}
	return nil
}

func (b *BBS) Version() (string, string, error) {
	var name string
	p, err := b.Get("post:0", SuperUser)
	fm := p.FrontMatter()
	if fm == nil {
		name = "unnamed"
	} else {
		name = fm.Title
	}
	return name, version, err
}

// allow checks if the user if able to read or write.
func (b *BBS) allow(key string, post Post, user string, write bool) bool {
	if user == SuperUser {
		return true
	}

	users := &SortedString{}
	check(b.meta("users", &users, nil))
	if !users.Contain(user) {
		return false
	}

	// deal with post
	parts := strings.Split(key, ":")
	if len(parts) < 2 {
		return false
	}

	switch parts[0] {
	case "post", "user":
		// always allow read
		if !write {
			return true
		}

		// new post
		if post.Rev == 0 {
			switch parts[0] {
			case "post":
				// check nextid
				var nextid int64
				b.meta("nextid", &nextid, nil)
				postid, err := strconv.ParseInt(parts[1], 16, 64)
				if err != nil {
					log.Println(err)
					return false
				}
				return postid < nextid && postkey(postid) == key
			case "user":
				return parts[1] == user
			}
		} else {
			// old post, check if in authors list
			fm := post.FrontMatter()
			if fm != nil {
				for _, a := range fm.Authors {
					if a == user || a == "*" {
						return true
					}
				}
			}
		}
	}

	// special case
	// always allow user to modify his profile
	if key == "user:"+user {
		return true
	}

	return false
}

// meta is modify with "bbs:" prefixed key.
func (b *BBS) meta(key string, value interface{}, mutate func(v interface{}) bool) error {
	return b.modify("bbs:"+key, value, mutate)
}

// modify can fetch the value of the key, optionally update it.
// if the update failed, mutate will be applied again.
func (b *BBS) modify(key string, value interface{}, mutate func(v interface{}) bool) error {
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
	if mutate == nil || !mutate(value) {
		return nil
	}

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
	check(b.meta("nextid", &nextid, func(v interface{}) bool {
		key = postkey(nextid)
		nextid++
		return true
	}))
	return key
}

func (b *BBS) NewUser(user string) error {
	var (
		users = &SortedString{}
		err   error
	)

	if err2 := b.meta("users", users, func(v interface{}) bool {
		ok := users.Insert(user)
		if !ok {
			err = ErrUserExists
		}
		return ok
	}); err2 != nil {
		return err2
	}

	return err
}

func (b *BBS) SetUserPass(user, pass string) error {
	var phrase string
	return b.modify("userpass:"+user, &phrase, func(v interface{}) bool {
		phrase = passhash(user, pass)
		return true
	})
}

func (b *BBS) AuthUserPass(user, pass string) bool {
	var phrase string
	if err := b.modify("userpass:"+user, &phrase, nil); err != nil {
		log.Println("Error authenticating:", err)
		return false
	}
	return (len(pass) == 0 && len(phrase) == 0) || (passhash(user, pass) == phrase)
}

func (b *BBS) init() {
	p, err := b.Get("post:0", SuperUser)
	check(err)
	if p.Rev == 0 {
		key := b.NewPostKey()
		if key != "post:0" {
			log.Fatal("Wrong post:0 key, ", key)
		}
		p.Rev++
		p.Content = []byte(`---
title: TagBBS
authors: [sysop]
tags: [sysop]
---

Welcome to TagBBS!
`)
		check(b.Put("post:0", p, SuperUser))
	}
}
