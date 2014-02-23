package auth

import (
	"bytes"
	"errors"
	"net/url"

	"code.google.com/p/go.crypto/ssh"
	"github.com/tagbbs/tagbbs/rkv"
)

var ErrNoMatchedPublicKey = errors.New("No Matched Public Key")

type PublicKey struct {
	// Usually this should be a scoped storage.
	Store rkv.Storage
}

func (p PublicKey) Set(user string, authorizedKeys []byte) error {
	v, err := p.Store.Get(user)
	if err != nil {
		return err
	}
	v.Rev++
	v.Content = authorizedKeys
	return p.Store.Put(user, v)
}

func (p PublicKey) Auth(params url.Values) (string, error) {
	user := params.Get("user")
	algo := params.Get("algo")
	pubkey := []byte(params.Get("pubkey"))

	v, err := p.Store.Get(user)
	if err != nil {
		return "", err
	}

	keys := v.Content
	for {
		var (
			pkey ssh.PublicKey
			ok   bool
		)
		pkey, _, _, keys, ok = ssh.ParseAuthorizedKey(keys)
		if !ok {
			return "", ErrNoMatchedPublicKey
		}
		if pkey.PublicKeyAlgo() == algo && bytes.Compare(ssh.MarshalPublicKey(pkey), pubkey) == 0 {
			return user, nil
		}
	}
}
