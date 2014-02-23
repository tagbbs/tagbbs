// A unified authentication module.
package auth

import (
	"errors"
	"net/url"
	"reflect"
)

var ErrAuthFailed = errors.New("Authentication Failed")

// Authentication just means knowing who you are.
type Authentication interface {
	// Auth should return with (username, err).
	Auth(url.Values) (string, error)
}

type AuthenticationList []Authentication

func (al AuthenticationList) Auth(params url.Values) (user string, err error) {
	for _, a := range al {
		user, err = a.Auth(params)
		if err == nil {
			return
		}
	}
	return "", ErrAuthFailed
}

func (al AuthenticationList) Of(t reflect.Type) Authentication {
	for _, a := range al {
		if reflect.TypeOf(a) == t {
			return a
		}
	}
	return nil
}
