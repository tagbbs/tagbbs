// Simple index upon the key value store.

package tagbbs

import (
	"bytes"
	"log"
	"strconv"
	"strings"
)

func analyze(key string, p Post) (keys []string) {
	switch {
	case strings.HasPrefix(key, "post:"):
		if len(bytes.Trim(p.Content, " \r\n\t")) == 0 {
			return
		}
		fm := p.FrontMatter()
		if fm == nil {
			return
		}
		keys = append(keys, fm.Tags...)
		if len(fm.Thread) > 0 {
			keys = append(keys, fm.Thread)
		} else {
			keys = append(keys, key)
		}
	}
	return
}

func (b *BBS) indexRemove(key string, p Post) {
	keywords := analyze(key, p)
	for _, word := range keywords {
		b.Storage.SetDelete("list:"+word, key)
	}
}

func (b *BBS) indexAdd(key string, p Post) {
	keywords := analyze(key, p)
	for _, word := range keywords {
		b.Storage.SetInsert("list:"+word, key)
	}
}

func (b *BBS) indexReplace(key string, oldpost, newpost Post) {
	b.indexRemove(key, oldpost)
	b.indexAdd(key, newpost)
}

type ParsedQuery struct {
	Tags   []string `json:"tags"`
	Cursor string   `json:"cursor"`
	Before int      `json:"before"`
	After  int      `json:"after"`
}

func parseQuery(q string) (p ParsedQuery) {
	parts := strings.Split(q, " ")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		switch part[0] {
		case '@':
			p.Cursor = part[1:]
		case '+':
			p.After, _ = strconv.Atoi(part[1:])
		case '-':
			p.Before, _ = strconv.Atoi(part[1:])
		default:
			p.Tags = append(p.Tags, part)
		}
	}
	return
}

func (b *BBS) Query(q string) (ss []string, p ParsedQuery, err error) {
	p = parseQuery(q)
	if len(p.Tags) == 0 {
		return []string{}, p, nil
	}
	if p.Before == 0 && p.After == 0 {
		p.Before = 20
	}
	ss, err = b.Storage.SetSlice("list:"+p.Tags[0], p.Cursor, p.Before, p.After)
	return
}

// Rebuild all index.
// Currently it should be called only when the database is exclusively locked.
func (b *BBS) RebuildIndex() {
	var checklog = func(err error) {
		if err != nil {
			log.Println(err)
		}
	}
	var i int64
	p, err := b.Get("bbs:nextid", SuperUser)
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
