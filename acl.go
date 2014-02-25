// Access Control.

package tagbbs

import (
	"log"
	"strconv"
	"strings"
)

func (b *BBS) NewUser(user string) error {
	var (
		users = &SortedString{}
		err   error
	)

	if err2 := b.modify("bbs:users", users, func(v interface{}) bool {
		ok := users.Insert(user)
		if !ok {
			err = ErrUserExists
		}
		return ok
	}); err2 != nil {
		return err2
	}

	return err
}

// allow checks if the user if able to read or write.
func (b *BBS) allow(key string, post Post, user string, write bool) bool {
	// SuperUser or in SYSOP List
	if user == SuperUser {
		return true
	} else {
		p, err := b.Get("user:"+SuperUser, SuperUser)
		if err != nil {
			return false
		}
		if fm := p.FrontMatter(); fm != nil {
			for _, a := range fm.Authors {
				if a == user {
					return true
				}
			}
		}
	}

	users := &SortedString{}
	check(b.modify("bbs:users", &users, nil))
	if !users.Contain(user) {
		return false
	}

	// deal with post
	parts := strings.Split(key, ":")
	if len(parts) < 2 {
		return false
	}

	switch parts[0] {
	case "post", "user":
		// always allow read
		if !write {
			return true
		}

		// new post
		if post.Rev == 0 {
			switch parts[0] {
			case "post":
				// check nextid
				p, _ := b.Get("bbs:nextid", SuperUser)
				nextid := p.Rev
				postid, err := strconv.ParseInt(parts[1], 16, 64)
				if err != nil {
					log.Println(err)
					return false
				}
				return postid < nextid && postkey(postid) == key
			case "user":
				return parts[1] == user
			}
		} else {
			// old post, check if in authors list
			fm := post.FrontMatter()
			if fm != nil {
				for _, a := range fm.Authors {
					if a == user || a == "*" {
						return true
					}
				}
			}
		}
	}

	// special case
	// always allow user to modify his profile
	if key == "user:"+user {
		return true
	}

	return false
}
