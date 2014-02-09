package tagbbs

import (
	"bytes"
	"testing"
)

func TestMemBBS(t *testing.T) {
	store := MemStore{}
	b := NewBBS(store)

	p1 := Post{}
	p1.Rev = 1
	p1.Content = []byte(`
---
title: Hello
tags:
  - a
  - b
---

Hello World!
`)
	// Test Put
	if err := b.Put(b.NewPostKey(), p1); err != nil {
		t.Fatal(err)
	}

	// Test Query
	var pid string
	if list, err := b.Query("a"); err != nil {
		t.Fatal(err)
	} else if len(list) != 1 {
		t.Fatal("Wrong number of posts returned.")
	} else {
		pid = list[0]
	}

	// Test Get
	if p2, err := b.Get(pid); err != nil {
		t.Fatal(err, pid)
	} else {
		if bytes.Compare(p1.Content, p2.Content) != 0 {
			t.Fatal("Content not match!")
		}
	}
}
