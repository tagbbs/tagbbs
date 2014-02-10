package tagbbs

import "errors"

var (
	ErrRevNotMatch   = errors.New("Revision Not Match")
	ErrNotLocked     = errors.New("Not Locked")
	ErrAlreadyLocked = errors.New("Already Locked")
)

type Storage interface {
	// Get the Post of given Key.
	Get(key string) (Post, error)
	// Set the Post to the given Key.
	// Note that the revision of the new post must be increased by 1 from the old post.
	// If the length of the Content of the Post is zero, the post is considered safe to be deleted.
	Put(key string, p Post) error
	// Enumerate all posts with given prefix.
	// Enumerate(prefix string, pp func(Post)) error
}
