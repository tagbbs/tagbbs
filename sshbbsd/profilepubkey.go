package main

import (
	"bytes"
	"net/url"

	"code.google.com/p/go.crypto/ssh"
	"github.com/tagbbs/tagbbs"
	"github.com/tagbbs/tagbbs/auth"
)

type ProfilePublicKeyAuth struct {
	*tagbbs.BBS
}

func (p ProfilePublicKeyAuth) Auth(params url.Values) (string, error) {
	user := params.Get("user")
	algo := params.Get("algo")
	pubkey := []byte(params.Get("pubkey"))

	type Profile struct {
		AuthorizedKeys string `yaml:"authorized_keys"`
	}
	post, err := p.Get("user:"+user, user)
	if err != nil {
		return "", err
	}
	profile := Profile{}
	err = post.UnmarshalTo(&profile)
	if err != nil {
		return "", err
	}
	keys := []byte(profile.AuthorizedKeys)
	for {
		var (
			pkey ssh.PublicKey
			ok   bool
		)
		pkey, _, _, keys, ok = ssh.ParseAuthorizedKey(keys)
		if !ok {
			return "", auth.ErrNoMatchedPublicKey
		}
		if pkey.PublicKeyAlgo() == algo && bytes.Compare(ssh.MarshalPublicKey(pkey), pubkey) == 0 {
			return user, nil
		}
	}
}
