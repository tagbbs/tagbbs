package tagbbs

import "sort"

type SortedString []string

func (s SortedString) Len() int      { return len(s) }
func (s SortedString) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortedString) Less(i, j int) bool {
	if len(s[i]) == len(s[j]) {
		return s[i] < s[j]
	}
	return len(s[i]) < len(s[j])
}
func (s SortedString) Sort() {
	sort.Sort(s)
}
func (s *SortedString) Unique() {
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
	if i := sort.StringSlice(*s).Search(val); i < len(*s) && (*s)[i] == val {
		return false
	} else {
		*s = append((*s)[:i], append([]string{val}, (*s)[i:]...)...)
		return true
	}
}
func (s *SortedString) Delete(val string) bool {
	if i := sort.StringSlice(*s).Search(val); i < len(*s) && (*s)[i] == val {
		*s = append((*s)[:i], (*s)[i+1:]...)
		return true
	} else {
		return false
	}
}
