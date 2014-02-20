package tagbbs

import (
	"bytes"
	"testing"

	"github.com/tagbbs/tagbbs/rkv"
)

func TestMemBBS(t *testing.T) {
	b := NewBBSFromString("mem://")

	p1 := Post{}
	p1.Rev = 1
	p1.Content = []byte(`
---
title: Hello
authors: [userA]
tags:
  - a
  - b
---

Hello World!
`)
	p1key := b.NewPostKey()
	p2key := b.NewPostKey()
	userA := "userA"
	userB := "userB"

	if err := b.NewUser(userA); err != nil {
		t.Fatal(err)
	}
	if err := b.NewUser(userB); err != nil {
		t.Fatal(err)
	}

	// Test Put
	if err := b.Put(p1key, p1, userA); err != nil {
		t.Fatal(err)
	}
	if err := b.Put(p2key, p1, userA); err != nil {
		t.Fatal(err)
	}
	// Test Revision
	if err := b.Put(p1key, p1, userA); err != rkv.ErrRevNotMatch {
		t.Fatal(err)
	}

	// Test Query
	var pid string
	if list, _, err := b.Query("a"); err != nil {
		t.Fatal(err)
	} else if len(list) != 2 {
		t.Fatal("Wrong number of posts returned.")
	} else {
		pid = list[0]
	}

	// Test Get
	if p2, err := b.Get(pid, userA); err != nil {
		t.Fatal(err, pid)
	} else {
		if bytes.Compare(p1.Content, p2.Content) != 0 {
			t.Fatal("Content not match!")
		}
	}

	// Test Permission
	if _, err := b.Get(p1key, userB); err != nil {
		t.Fatal(err, pid)
	}
	if _, err := b.Get(p1key, "NotExistedUser"); err != ErrAccessDenied {
		t.Fatal(err)
	}
	if err := b.Put(p1key, p1, userB); err != ErrAccessDenied {
		t.Fatal(err)
	}

	// Remove
	p1.Rev++
	p1.Content = []byte("")
	if err := b.Put(p1key, p1, SuperUser); err != nil {
		t.Fatal(err)
	}

	// Test Query Again
	if list, _, err := b.Query("a"); err != nil {
		t.Fatal(err)
	} else if len(list) != 1 {
		t.Fatal("Wrong number of posts returned.", list)
	}
}
