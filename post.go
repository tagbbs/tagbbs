package tagbbs

import (
	"bytes"
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
}

var separator = []byte("---\n")

func (p *Post) sep() ([]byte, []byte) {
	con := bytes.TrimLeft(p.Content, "\r\n\t ")
	frontBegin := bytes.Index(con, separator)
	if frontBegin != 0 {
		return nil, con
	}
	frontBegin += len(separator)
	frontLength := bytes.Index(con[frontBegin:], separator)
	if frontLength < 0 {
		return nil, con
	}
	return con[frontBegin : frontBegin+frontLength], con[frontBegin+frontLength+len(separator):]
}

func (p *Post) FrontMatter() *FrontMatter {
	fmb, _ := p.sep()
	if len(fmb) == 0 {
		return nil
	}
	var fm FrontMatter
	if err := goyaml.Unmarshal(fmb, &fm); err != nil {
		return nil
	}
	return &fm
}

func (p *Post) Body() []byte {
	_, body := p.sep()
	return body
}
