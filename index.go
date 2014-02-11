package tagbbs

import (
	"sort"
)

func (b *BBS) indexRemove(key string, p Post) {
	fm := p.FrontMatter()
	if fm == nil {
		return
	}
	tags := []string{""}
	tags = append(tags, fm.Tags...)
	for _, tag := range tags {
		var ids []string
		b.modify("tag:"+tag, &ids, func(v interface{}) bool {
			i := sort.StringSlice(ids).Search(key)
			if i < len(ids) && ids[i] == key {
				ids = append(ids[:i], ids[i+1:]...)
			}
			return true
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
		var ids []string
		b.modify("tag:"+tag, &ids, func(v interface{}) bool {
			i := sort.StringSlice(ids).Search(key)
			if i >= len(ids) || ids[i] != key {
				ids = append(ids[:i], append([]string{key}, ids[i:]...)...)
			}
			return true
		})
	}
}

func (b *BBS) Query(q string) ([]string, error) {
	var ids []string
	err := b.modify("tag:"+q, &ids, nil)
	return ids, err
}
