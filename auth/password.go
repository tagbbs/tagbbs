package auth

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/tagbbs/tagbbs/rkv"
)

var ErrWrongPassword = errors.New("Wrong Password")

type Password struct {
	// Usually this should be a scoped storage.
	Store rkv.Storage
}

func (p Password) Set(user, pass string) error {
	v, err := p.Store.Get(user)
	if err != nil {
		return err
	}
	v.Rev++
	v.Content, err = json.Marshal(passhash(user, pass))
	if err != nil {
		return err
	}
	return p.Store.Put(user, v)
}

// v is a map with keys "user" and "pass".
func (p Password) Auth(params url.Values) (string, error) {
	user, pass := params.Get("user"), params.Get("pass")
	v, err := p.Store.Get(user)
	if err != nil {
		return "", err
	}
	var phrase string
	json.Unmarshal(v.Content, &phrase)
	if phrase != passhash(user, pass) {
		return "", ErrAuthFailed
	}
	return user, nil
}

// XXX need some improvements
func passhash(user, pass string) string {
	h := crypto.SHA256.New()
	h.Write([]byte("TESTING" + user + "|" + pass + "|" + user + "TESTING"))
	x := h.Sum(nil)
	return hex.EncodeToString(x)
}
