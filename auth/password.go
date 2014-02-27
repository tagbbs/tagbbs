package auth

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/tagbbs/tagbbs/rkv"
)

var (
	ErrUserExist     = errors.New("User Exist")
	ErrUserNotExist  = errors.New("User Not Exist")
	ErrWrongPassword = errors.New("Wrong Password")
)

type Password struct {
	// Usually this should be a scoped storage.
	Store rkv.Interface
}

func (p Password) New(user, pass string) error {
	v, err := p.Store.Get(user)
	if err != nil {
		return err
	}
	if v.Rev != 0 {
		return ErrUserExist
	}
	v.Rev++
	v.Content, err = json.Marshal(passhash(user, pass))
	if err != nil {
		return err
	}
	return p.Store.Put(user, v)
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
	if v.Rev == 0 {
		return "", ErrUserNotExist
	}
	var phrase string
	if err := json.Unmarshal(v.Content, &phrase); err == nil {
		if phrase == passhash(user, pass) {
			return user, nil
		}
	}
	return "", ErrAuthFailed
}

// XXX need some improvements
func passhash(user, pass string) string {
	h := crypto.SHA256.New()
	h.Write([]byte("TESTING" + user + "|" + pass + "|" + user + "TESTING"))
	x := h.Sum(nil)
	return hex.EncodeToString(x)
}
