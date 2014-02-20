package main

import (
	"flag"

	_ "github.com/go-sql-driver/mysql"
	"github.com/thinxer/tagbbs"
)

var flagDB = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")

func main() {
	flag.Parse()
	bbs := tagbbs.NewBBSFromString(*flagDB)
	bbs.RebuildIndex()
}
