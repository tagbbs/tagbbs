package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tagbbs/tagbbs"
)

var (
	flagDB         = flag.String("db", "mem://", "connection string")
	flagCpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")
	flagMemProfile = flag.String("memprofile", "", "write memory profile to this file")
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
		err := bbs.Put(key, p, "u1")
		if err != nil {
			log.Panicln(key, err)
		}
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
	if *flagCpuProfile != "" {
		f, err := os.Create(*flagCpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	bbs := tagbbs.NewBBSFromString(*flagDB)
	start := time.Now()

	threads := 10
	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go batchput(bbs)
	}
	wg.Wait()

	dur := time.Now().Sub(start)
	log.Println("All Finished, total: ", allCount)
	log.Println("Duration: ", dur, ", Average: ", dur/time.Duration(allCount))

	batchget(bbs)

	if *flagMemProfile != "" {
		f, err := os.Create(*flagMemProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}
}
