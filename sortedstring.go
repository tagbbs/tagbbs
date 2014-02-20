package tagbbs

import "sort"

// SortedString is a list of sorted strings, mainly for indexing purpose.
// Note that the order of strings are defined by (len, content).
type SortedString []string

func strcmp(a, b string) bool {
	if len(a) == len(b) {
		return a < b
	}
	return len(a) < len(b)

}

func (s SortedString) Len() int           { return len(s) }
func (s SortedString) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortedString) Less(i, j int) bool { return strcmp(s[i], s[j]) }
func (s SortedString) Sort() {
	sort.Sort(s)
}
func (s SortedString) Search(v string) int {
	return sort.Search(len(s), func(i int) bool { return !strcmp(s[i], v) })
}
func (s SortedString) Contain(v string) bool {
	i := s.Search(v)
	return i < len(s) && s[i] == v
}
func (s SortedString) norm(i int) int {
	if i < 0 {
		return 0
	}
	if i > len(s) {
		return len(s)
	}
	return i
}
func (s SortedString) Slice(v string, before int, after int) SortedString {
	if len(s) == 0 {
		return SortedString{}
	}
	i := s.Search(v)
	if i == 0 && s[i] != v && before > 0 {
		i = len(s)
	}
	before = s.norm(i - before)
	after = s.norm(i + after)
	if before > after {
		after = before
	}
	return SortedString(s[before:after])
}
func (s *SortedString) Unique() {
	if len(*s) == 0 {
		return
	}
	i := 0
	for j := 1; j < len(*s); j++ {
		if (*s)[i] != (*s)[j] {
			i++
		}
		(*s)[i] = (*s)[j]
	}
	(*s) = (*s)[:i+1]
}
func (s *SortedString) Insert(val string) bool {
	if i := s.Search(val); i < len(*s) && (*s)[i] == val {
		return false
	} else {
		*s = append((*s)[:i], append([]string{val}, (*s)[i:]...)...)
		return true
	}
}
func (s *SortedString) Delete(val string) bool {
	if i := s.Search(val); i < len(*s) && (*s)[i] == val {
		*s = append((*s)[:i], (*s)[i+1:]...)
		return true
	} else {
		return false
	}
}

func SortedUnion(sss ...SortedString) (ret SortedString) {
	for _, ss := range sss {
		ret = append(ret, ss...)
	}
	ret.Sort()
	ret.Unique()
	return
}

func SortedIntersect(sss ...SortedString) (ret SortedString) {
	if len(sss) == 0 {
		return
	}
	ret = make(SortedString, len(sss[0]))
	copy(ret, sss[0])
	for _, ss := range sss[1:] {
		for _, s := range ret {
			if !ss.Contain(s) {
				ret.Delete(s)
			}
		}
	}
	return
}
