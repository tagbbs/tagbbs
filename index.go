package tagbbs

import (
	"strconv"
	"strings"
)

func (b *BBS) indexRemove(key string, p Post) {
	fm := p.FrontMatter()
	if fm == nil {
		return
	}
	tags := []string{""}
	tags = append(tags, fm.Tags...)
	for _, tag := range tags {
		ids := &SortedString{}
		b.modify("tag:"+tag, &ids, func(v interface{}) bool {
			ids.Sort() // XXX temporary fix
			return ids.Delete(key)
		})
	}
}

func (b *BBS) indexAdd(key string, p Post) {
	fm := p.FrontMatter()
	if fm == nil {
		return
	}
	tags := []string{""}
	tags = append(tags, fm.Tags...)
	for _, tag := range tags {
		ids := &SortedString{}
		b.modify("tag:"+tag, &ids, func(v interface{}) bool {
			ids.Sort() // XXX temporary fix
			return ids.Insert(key)
		})
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

func (b *BBS) Query(q string) ([]string, ParsedQuery, error) {
	p := parseQuery(q)
	if len(p.Tags) == 0 {
		p.Tags = []string{""}
	}
	if p.Before == 0 && p.After == 0 {
		p.Before = 20
	}
	var ids SortedString
	err := b.modify("tag:"+p.Tags[0], &ids, nil)
	ids = ids.Slice(p.Cursor, p.Before, p.After)
	return ids, p, err
}
