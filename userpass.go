package tagbbs

import (
	"crypto"
	"encoding/hex"
	"log"
)

func (b *BBS) SetUserPass(user, pass string) error {
	var phrase string
	return b.modify("userpass:"+user, &phrase, func(v interface{}) bool {
		phrase = passhash(user, pass)
		return true
	})
}

func (b *BBS) AuthUserPass(user, pass string) bool {
	var phrase string
	if err := b.modify("userpass:"+user, &phrase, nil); err != nil {
		log.Println("Error authenticating:", err)
		return false
	}
	return (len(pass) == 0 && len(phrase) == 0) || (passhash(user, pass) == phrase)
}

// XXX need some improvements
func passhash(user, pass string) string {
	h := crypto.SHA256.New()
	h.Write([]byte("TESTING" + user + "|" + pass + "|" + user + "TESTING"))
	x := h.Sum(nil)
	return hex.EncodeToString(x)
}
