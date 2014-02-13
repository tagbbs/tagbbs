package tagbbs

import (
	"strings"
	"testing"
)

func TestSortedString(t *testing.T) {
	s := &SortedString{}
	*s = append(*s, strings.Split("d b b a c a d", " ")...)
	s.Sort()
	s.Unique()
	if s.Len() != 4 {
		t.Fail()
	}
}
