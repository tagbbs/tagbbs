package tagbbs

import (
	"fmt"
	"time"

	"launchpad.net/goyaml"
)

type Post struct {
	Rev       int64
	Timestamp time.Time
	Content   []byte
}

type FrontMatter struct {
	Title   string
	Tags    []string
	Authors []string
	Reply   string
	Thread  string
}

func (p *Post) FrontMatter() *FrontMatter {
	var fm FrontMatter
	if err := goyaml.Unmarshal(p.Content, &fm); err != nil {
		return nil
	}
	return &fm
}

func (p *Post) UnmarshalTo(v interface{}) error {
	return goyaml.Unmarshal(p.Content, v)
}

func (p *Post) String() string {
	return fmt.Sprintf("%d %v %s", p.Rev, p.Timestamp, string(p.Content))
}
