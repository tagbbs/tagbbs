package tagbbs

import (
	"errors"
	"strings"
)

var (
	ErrRevNotMatch = errors.New("Revision Not Match")
)

type Storage interface {
	// Get the Post of given Key.
	Get(key string) (Post, error)
	// Set the Post to the given Key.
	// Note that the revision of the new post must be increased by 1 from the old post.
	// If the length of the Content of the Post is zero, the post is considered safe to be deleted.
	Put(key string, p Post) error
}

// NewStore is a helper method for creating a built-in storage.
func NewStore(source string) (store Storage, err error) {
	parts := strings.SplitN(source, "://", 2)
	driver := parts[0]
	if driver == "mysql" {
		store, err = NewSQLStore(driver, parts[1], "kvs")
	} else if driver == "mem" {
		store = MemStore{}
	} else {
		panic("unknown driver: " + driver)
	}
	return store, err
}
