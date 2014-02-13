package tagbbs

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

func (b *BBS) Query(q string) ([]string, error) {
	var ids []string
	err := b.modify("tag:"+q, &ids, nil)
	return ids, err
}
