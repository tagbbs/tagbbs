package tagbbs

import (
	"sort"
)

func (b *BBS) indexRemove(key string, p Post) {
	fm := p.FrontMatter()
	if fm == nil {
		return
	}
	for _, tag := range fm.Tags {
		var ids []string
		b.modify("tag:"+tag, &ids, func(v interface{}) {
			i := sort.StringSlice(ids).Search(key)
			if i < len(ids) && ids[i] == key {
				ids = append(ids[:i], ids[i+1:]...)
			}
		})
	}
}

func (b *BBS) indexAdd(key string, p Post) {
	fm := p.FrontMatter()
	if fm == nil {
		return
	}
	for _, tag := range fm.Tags {
		var ids []string
		b.modify("tag:"+tag, &ids, func(v interface{}) {
			i := sort.StringSlice(ids).Search(key)
			if i >= len(ids) || ids[i] != key {
				ids = append(ids[:i], append([]string{key}, ids[i:]...)...)
			}
		})
	}
}

func (b *BBS) Query(q string) ([]string, error) {
	var ids []string
	err := b.modify("tag:"+q, &ids, nil)
	return ids, err
}
