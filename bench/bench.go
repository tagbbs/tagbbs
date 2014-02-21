package main

import (
	"flag"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tagbbs/tagbbs"
)

var (
	flagDB = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")
)

var SAMPLE_POST = []byte(`
---
title: SAMPLE_POST
authors: [u1, u2]
tags: [test]
---
TEST
`)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var wg sync.WaitGroup
var allCount int64

func batchput(bbs *tagbbs.BBS) {
	start := time.Now()
	log.Println("Start")
	count := 1000
	for i := 0; i < count; i++ {
		key := bbs.NewPostKey()
		p := tagbbs.Post{}
		p.Rev = 1
		p.Content = SAMPLE_POST
		bbs.Put(key, p, "u1")
	}
	dur := time.Now().Sub(start)
	log.Println("Duration: ", dur, ", Average: ", dur/time.Duration(count))
	atomic.AddInt64(&allCount, int64(count))
	wg.Done()
}

func batchget(bbs *tagbbs.BBS) {
	p, e := bbs.Get("bbs:nextid", tagbbs.SuperUser)
	check(e)
	count := int(p.Rev)
	log.Println("Start for ", count)
	start := time.Now()
	for i := 0; i < count; i++ {
		key := "post:" + strconv.FormatInt(int64(i), 16)
		_, e := bbs.Get(key, "u2")
		check(e)
	}
	dur := time.Now().Sub(start)
	log.Println("Duration: ", dur, ", Average: ", dur/time.Duration(count))
}

func main() {
	flag.Parse()

	bbs := tagbbs.NewBBSFromString(*flagDB)
	bbs.NewUser("u1")
	bbs.NewUser("u2")

	start := time.Now()

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go batchput(bbs)
	}
	wg.Wait()

	dur := time.Now().Sub(start)
	log.Println("All Finished, total: ", allCount)
	log.Println("Duration: ", dur, ", Average: ", dur/time.Duration(allCount))

	batchget(bbs)
}
