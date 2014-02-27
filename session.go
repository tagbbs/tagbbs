package tagbbs

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/tagbbs/tagbbs/rkv"
)

var ErrSessionNotExist = errors.New("Session Not Exist")

type Capability int64

const (
	CapRead Capability = 1 << iota
	CapPost
	CapNotification
	CapChat
)

type SessionManager struct {
	Store rkv.S
}

type Session struct {
	User       string `json:"user"`
	UserAgent  string `json:"user_agent"`
	Capability `json:"capability"`
	LoginTime  time.Time `json:"login_time"`
	RemoteAddr string    `json:"remote_addr"`
}

func (s SessionManager) Request(ss Session) (sid string, err error) {
	randbits := make([]byte, 16)
	_, err = rand.Read(randbits)
	if err != nil {
		return
	}
	sid = hex.EncodeToString(randbits)

	ss.LoginTime = time.Now()
	err = s.Store.Write("session:"+sid, 1, ss)
	if err != nil {
		return
	}

	err = s.Store.SetInsert("usersessions:"+ss.User, sid)
	return
}

func (s SessionManager) Revoke(sid string) (err error) {
	var session Session
	var user string
	err = s.Store.ReadModify("session:"+sid, &session, func(_ interface{}) bool {
		user = session.User
		session = Session{}
		return true
	})
	if err != nil {
		return
	}
	err = s.Store.SetDelete("usersessions:"+user, sid)
	return
}

func (s SessionManager) Get(sid string) (session Session, err error) {
	err = s.Store.Read("session:"+sid, nil, &session)
	if len(session.User) == 0 {
		err = ErrSessionNotExist
	}
	return
}

func (s SessionManager) SetCapability(sid string, capability Capability) error {
	var session Session
	return s.Store.ReadModify("session:"+sid, &session, func(_ interface{}) bool {
		session.Capability = capability
		return true
	})
}

func (s SessionManager) List(user string) (ss map[string]Session, err error) {
	var sss []string
	sss, err = s.Store.SetSlice("usersessions:"+user, "", 64, 0)
	if err != nil {
		return
	}
	ss = make(map[string]Session)
	for _, sid := range sss {
		var session Session
		session, err = s.Get(sid)
		if err != nil {
			return
		}
		ss[sid] = session
	}
	return
}
