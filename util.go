package tagbbs

import "strconv"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func postkey(id int64) string {
	return "post:" + strconv.FormatInt(id, 16)
}
