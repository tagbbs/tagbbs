package main

import (
	"flag"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tagbbs/tagbbs"
)

var flagDB = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")

func main() {
	flag.Parse()
	bbs := tagbbs.NewBBSFromString(*flagDB)
	bbs.RebuildIndex()
}

// Rebuild all index.
// Currently it should be called only when the database is exclusively locked.
func RebuildIndex(b) {
	var checklog = func(err error) {
		if err != nil {
			log.Println(err)
		}
	}
	var i int64
	p, err := in.Get("bbs:nextid")
	if err != nil {
		panic(err)
	}
	nextid := p.Rev
	log.Println("NextId:", nextid)
	// gather tags
	var allNames map[string]bool
	for i = 0; i < nextid; i++ {
		key := postkey(i)
		p, err := b.Get(key, SuperUser)
		checklog(err)
		names := analyze(key, p)
		for _, name := range names {
			allNames[name] = true
		}
	}
	log.Println("All List Names:", len(allNames))
	// remove all lists
	for name := range allNames {
		var tmp []string
		checklog(b.Storage.ReadModify("list:"+name, &tmp, func(v interface{}) bool {
			tmp = nil
			return true
		}))
	}
	// re-add all posts
	var count int64
	for i = 0; i < nextid; i++ {
		key := postkey(i)
		p, err := b.Get(key, SuperUser)
		checklog(err)
		b.indexReplace(key, Post{}, p)
		if len(p.Content) != 0 {
			count++
		}
	}
	log.Printf("Total: %v posts, %v tags\n", count, len(allNames))
}
