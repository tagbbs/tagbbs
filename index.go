package tagbbs

type idlist []string

func (b *BBS) indexRemove(key string, p Post) {
	fm := p.FrontMatter()
	if fm == nil {
		return
	}
	for _, tag := range fm.Tags {
		var ids idlist
		b.modify("tag:"+tag, &ids, func(v interface{}) {
			for i, id := range ids {
				if id == key {
					ids = append(ids[:i], ids[i+1:]...)
					break
				}
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
		var ids idlist
		b.modify("tag:"+tag, &ids, func(v interface{}) {
			ids = append(ids, key)
		})
	}
}

func (b *BBS) Query(q string) ([]string, error) {
	var ids idlist
	err := b.modify("tag:"+q, &ids, nil)
	return ids, err
}
