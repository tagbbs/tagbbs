// This is the core library for tagbbs, which will be used across different servers.
package tagbbs

import (
	"io/ioutil"
	"log"
	"launchpad.net/goyaml"

	"github.com/tagbbs/tagbbs/auth"
	"github.com/tagbbs/tagbbs/rkv"
)

const (
	SuperUser = "sysop"
)

var version string

func init() {
	if len(version) == 0 {
		version = "dev"
	}
}

type BBS struct {
	Storage rkv.S
	Index
	Auth auth.AuthenticationList
	SessionManager
}

func NewBBSFromConfig(configFile string) *BBS {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	var config struct {
		Storage string
		Index   struct {
			Storage string
		}
		Session struct {
			Storage string
		}
		// TODO make this more flexible
		Auth []struct {
			Type    string
			Storage string
		}
	}
	if err := goyaml.Unmarshal(bytes, &config); err != nil {
		panic(err)
	}

	must := func(s rkv.Interface, err error) rkv.Interface {
		if err != nil {
			panic(err)
		}
		return s
	}

	bbs := &BBS{
		Storage: rkv.S{must(rkv.NewStore(config.Storage))},
		Index: Index{
			rkv.S{
				rkv.ScopedStore{must(rkv.NewStore(config.Index.Storage)), "list:"}},
		},
		SessionManager: SessionManager{
			rkv.S{
				rkv.ScopedStore{must(rkv.NewStore(config.Session.Storage)), "session:"},
			},
		},
	}
	alist := auth.AuthenticationList{}
	for _, au := range config.Auth {
		switch au.Type {
		case "Password":
			alist = append(alist, auth.Password{
				rkv.ScopedStore{must(rkv.NewStore(au.Storage)), "userpass:"},
			})
		}
	}
	bbs.Auth = alist

	return bbs
}

func (b *BBS) Init() {
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

	name, version, err := b.Version()
	if err != nil {
		panic(err)
	}
	log.Println("Initialized:", name, version)
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
	b.Index.Replace(key, Post(oldpost), Post(post))
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
