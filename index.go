// Simple index upon the key value store.

package tagbbs

import (
	"bytes"
	"log"
	"strconv"
	"strings"
)

func analyze(key string, p Post) (keys SortedString) {
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
		keys.Sort()
		keys.Unique()
	}
	return
}

func (b *BBS) indexRemove(key string, p Post) SortedString {
	names := analyze(key, p)
	for _, name := range names {
		ids := &SortedString{}
		b.modify("list:"+name, &ids, func(v interface{}) bool {
			return ids.Delete(key)
		})
	}
	return names
}

func (b *BBS) indexAdd(key string, p Post) SortedString {
	keywords := analyze(key, p)
	for _, word := range keywords {
		ids := &SortedString{}
		b.modify("list:"+word, &ids, func(v interface{}) bool {
			return ids.Insert(key)
		})
	}
	return keywords
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

func (b *BBS) Query(q string) ([]string, ParsedQuery, error) {
	p := parseQuery(q)
	if len(p.Tags) == 0 {
		p.Tags = []string{""}
	}
	if p.Before == 0 && p.After == 0 {
		p.Before = 20
	}
	var lists []SortedString
	for _, tag := range p.Tags {
		var ids SortedString
		err := b.modify("list:"+tag, &ids, nil)
		if err != nil {
			return nil, p, err
		}
		lists = append(lists, ids)
	}
	ids := SortedIntersect(lists...)
	ids = ids.Slice(p.Cursor, p.Before, p.After)
	return ids, p, nil
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
	var allNames, tmp SortedString
	for i = 0; i < nextid; i++ {
		key := postkey(i)
		p, err := b.Get(key, SuperUser)
		checklog(err)
		names := analyze(key, p)
		allNames = append(allNames, names...)
		allNames.Sort()
		allNames.Unique()
	}
	log.Println("All List Names:", len(allNames))
	// remove all lists
	for _, name := range allNames {
		checklog(b.modify("list:"+name, &tmp, func(v interface{}) bool {
			tmp = SortedString{}
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
