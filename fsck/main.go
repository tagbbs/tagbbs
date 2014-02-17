package main

import (
	"flag"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/thinxer/tagbbs"
)

var flagDB = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")

func main() {
	flag.Parse()

	store, err := tagbbs.NewStore(*flagDB)
	if err != nil {
		panic(err)
	}
	bbs := tagbbs.NewBBS(store)
	log.Println(bbs.Version())
	bbs.RebuildIndex()
}
