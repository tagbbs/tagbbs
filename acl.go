// Access Control.

package tagbbs

import (
	"log"
	"strconv"
	"strings"
)

// allow checks if the user if able to read or write.
func (b *BBS) allow(key string, post Post, user string, write bool) bool {
	// Special Case: always allow SuperUsers
	// SuperUser or in SuperUser's author list
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
