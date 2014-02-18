package tagbbs

import (
	"crypto"
	"encoding/hex"
	"strconv"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// XXX need some improvements
func passhash(user, pass string) string {
	h := crypto.SHA256.New()
	h.Write([]byte("TESTING" + user + "|" + pass + "|" + user + "TESTING"))
	x := h.Sum(nil)
	return hex.EncodeToString(x)
}

func postkey(id int64) string {
	return "post:" + strconv.FormatInt(id, 16)
}
