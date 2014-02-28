// Simple index upon the key value store.

package tagbbs

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/tagbbs/tagbbs/rkv"
)

type Index struct {
	rkv.S
}

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

func (in Index) Remove(key string, p Post) {
	keywords := analyze(key, p)
	for _, word := range keywords {
		in.SetDelete(word, key)
	}
}

func (in Index) Add(key string, p Post) {
	keywords := analyze(key, p)
	for _, word := range keywords {
		in.SetInsert(word, key)
	}
}

func (in Index) Replace(key string, oldpost, newpost Post) {
	in.Remove(key, oldpost)
	in.Add(key, newpost)
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

func (in Index) Query(q string) (ss []string, p ParsedQuery, err error) {
	p = parseQuery(q)
	if len(p.Tags) == 0 {
		return []string{}, p, nil
	}
	if p.Before == 0 && p.After == 0 {
		p.Before = 20
	}
	ss, err = in.SetSlice(p.Tags[0], p.Cursor, p.Before, p.After)
	return
}
