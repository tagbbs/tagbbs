/*
Key Value Storage with Revision
*/
package rkv

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrRevNotMatch = errors.New("Revision Not Match")
)

type Value struct {
	Rev       int64
	Timestamp time.Time
	Content   []byte
}

type Storage interface {
	// Get the Post of given Key.
	Get(key string) (Value, error)
	// Set the Post to the given Key.
	// Note that the revision of the new post must be increased by 1 from the old post.
	// If the length of the Content of the Post is zero, the post is considered safe to be deleted.
	Put(key string, p Value) error
}

// NewStore is a helper method for creating a built-in storage.
func NewStore(source string) (store Storage, err error) {
	parts := strings.SplitN(source, "://", 2)
	driver := parts[0]
	if driver == "mysql" {
		store, err = NewSQLStore(driver, parts[1], "kvs")
	} else if driver == "redis" {
		store, err = NewRediStore(parts[1])
	} else if driver == "mem" {
		store = MemStore{}
	} else {
		panic("unknown driver: " + driver)
	}
	return store, err
}
