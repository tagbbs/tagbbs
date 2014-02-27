// This is the core library for tagbbs, which will be used across different servers.
package tagbbs

import (
	"errors"
	"log"

	"github.com/tagbbs/tagbbs/rkv"
)

var (
	ErrAccessDenied = errors.New("Access Denied")
)

type BBS struct {
	Storage rkv.S
	SessionManager
}

func NewBBS(store rkv.Interface) *BBS {
	b := &BBS{
		Storage: rkv.S{store},
		SessionManager: SessionManager{
			rkv.S{rkv.ScopedStore{store, "session:"}},
		},
	}
	b.init()
	name, version, err := b.Version()
	if err != nil {
		panic(err)
	}
	log.Println(name, version)
	return b
}

func NewBBSFromString(db string) *BBS {
	store, err := rkv.NewStore(db)
	check(err)
	return NewBBS(store)
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

// Version will return the name and build version of this installation.
// name is determined by the title of post:0,
// and build version is determined at compilation.
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

// Get will return the post requested by the user, if the user is permitted.
func (b *BBS) Get(key string, user string) (Post, error) {
	p, err := b.Storage.Get(key)
	if err != nil {
		return Post{}, err
	}
	if !b.allow(key, Post(p), user, false) {
		return Post{}, ErrAccessDenied
	}
	return Post(p), err
}

// Put will update the post and the index, if the user is permitted.
func (b *BBS) Put(key string, post Post, user string) error {
	oldpost, err := b.Storage.Get(key)
	if err != nil {
		return err
	}
	if !b.allow(key, Post(oldpost), user, true) {
		return ErrAccessDenied
	}
	if err := b.Storage.Put(key, rkv.Value(post)); err != nil {
		return err
	}
	b.indexReplace(key, Post(oldpost), Post(post))
	return nil
}

// Atomically obtain the next available key for post.
func (b *BBS) NewPostKey() string {
	for {
		p, err := b.Get("bbs:nextid", SuperUser)
		check(err)
		nextid := p.Rev
		p.Rev++
		err = b.Put("bbs:nextid", p, SuperUser)
		if err == nil {
			return postkey(nextid)
		} else if err == rkv.ErrRevNotMatch {
			continue
		} else {
			panic(err)
		}
	}
}
