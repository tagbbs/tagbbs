/*
Key Value Storage with Revision
*/
package rkv

import (
	"errors"
	"net/url"
	"strconv"
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
	u, err := url.Parse(source)
	if err != nil {
		return
	}

	switch u.Scheme {
	case "mysql":
		store, err = NewSQLStore(u.Scheme, u.User.String()+"@"+u.Host+u.RequestURI(), "kvs")
	case "redis":
		var db int
		if len(u.Path) > 0 {
			db, err = strconv.Atoi(u.Path[1:])
		}
		store, err = NewRediStore(u.Host, db)
	case "mem":
		store = NewMemStore()
	default:
		panic("unknown driver: " + u.Scheme)
	}
	return store, err
}
