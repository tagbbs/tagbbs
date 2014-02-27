package main

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/tagbbs/tagbbs"
	"github.com/tagbbs/tagbbs/auth"
	"github.com/tagbbs/tagbbs/rkv"
)

var (
	bbs   *tagbbs.BBS
	auths auth.AuthenticationList
)

func bbsinit() {
	bbs = tagbbs.NewBBSFromString(*flagDB)
	auths = auth.AuthenticationList{
		auth.Password{rkv.ScopedStore{bbs.Storage, "userpass:"}},
	}

	http.HandleFunc("/login", api(login, false))
	http.HandleFunc("/logout", api(logout, false))
	http.HandleFunc("/version", api(version, false))
	http.HandleFunc("/register", api(register, false))
	http.HandleFunc("/sessions", api(sessions, true))
	http.HandleFunc("/passwd", api(passwd, true))
	http.HandleFunc("/who", api(who, true))
	http.HandleFunc("/list", api(list, true))
	http.HandleFunc("/get", api(get, true))
	http.HandleFunc("/put", api(put, true))
}

func version(_, _ string, params url.Values) (result interface{}, err error) {
	var name, ver string
	name, ver, err = bbs.Version()
	result = M{
		"name":    name,
		"version": ver,
	}
	return
}

func login(_, _ string, params url.Values) (sid interface{}, err error) {
	var user string
	user, err = auths.Auth(params)
	if err != nil {
		return
	}
	sid, err = bbs.SessionManager.Request(tagbbs.Session{
		User:       user,
		UserAgent:  "Browser",
		Capability: tagbbs.CapRead | tagbbs.CapPost,
		RemoteAddr: params.Get("REMOTE_ADDR"),
	})
	if err != nil {
		return
	}
	return
}

func logout(_, _ string, params url.Values) (interface{}, error) {
	return nil, bbs.SessionManager.Revoke(params.Get("session"))
}

func register(_, _ string, params url.Values) (interface{}, error) {
	pw := auths.Of(reflect.TypeOf(auth.Password{})).(auth.Password)
	return nil, pw.New(params.Get("user"), params.Get("pass"))
}

func who(api, user string, params url.Values) (interface{}, error) {
	return user, nil
}

func sessions(api, user string, params url.Values) (interface{}, error) {
	return bbs.SessionManager.List(user)
}

func passwd(api, user string, params url.Values) (interface{}, error) {
	pass := params.Get("pass")
	pw := auths.Of(reflect.TypeOf(auth.Password{})).(auth.Password)
	return nil, pw.Set(user, pass)
}

func list(api, user string, params url.Values) (interface{}, error) {
	ids, parsed, err := bbs.Query(params.Get("query"))
	if err != nil {
		return nil, err
	}
	r := V{}
	for _, id := range ids {
		post, _ := bbs.Get(id, user)
		fm := post.FrontMatter()
		if fm == nil {
			continue
		}
		r = append(r, M{
			"key":     id,
			"title":   fm.Title,
			"authors": fm.Authors,
			"tags":    fm.Tags,
		})
	}
	return M{
		"posts": r,
		"query": parsed,
	}, err
}

func get(api, user string, params url.Values) (interface{}, error) {
	key := params.Get("key")
	post, err := bbs.Get(key, user)
	if err != nil {
		return nil, err
	}
	return M{
		"rev":       post.Rev,
		"timestamp": post.Timestamp,
		"content":   string(post.Content),
	}, nil
}

func put(api, user string, params url.Values) (interface{}, error) {
	var (
		err  error
		post tagbbs.Post
	)

	key := params.Get("key")
	if len(key) == 0 {
		key = bbs.NewPostKey()
	} else if post, err = bbs.Get(key, user); err != nil {
		return nil, err
	}
	post.Rev, _ = strconv.ParseInt(params.Get("rev"), 10, 64)
	post.Content = []byte(params.Get("content"))
	err = bbs.Put(key, post, user)
	if err == nil {
		return key, nil
	} else {
		return nil, err
	}
}
