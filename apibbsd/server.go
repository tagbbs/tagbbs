package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/thinxer/tagbbs"
)

var (
	flagDB     = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")
	flagListen = flag.String("listen", ":8023", "address to listen on")

	bbs *tagbbs.BBS
)

var (
	ErrAuthFailed   = errors.New("Authentication Failed")
	ErrUnauthorized = errors.New("Unauthorized")
)

var (
	sessions      = make(map[string]string)
	sessionsMutex sync.Mutex
)

type V []interface{}
type M map[string]interface{}
type apiHandler func(api, user string, params url.Values) (interface{}, error)

func bbsinit() {
	var (
		store tagbbs.Storage
		err   error
	)
	parts := strings.SplitN(*flagDB, "://", 2)
	driver := parts[0]
	if driver == "mysql" {
		store, err = tagbbs.NewSQLStore(driver, parts[1], "kvs")
		if err != nil {
			panic(err)
		}
	} else if driver == "mem" {
		store = tagbbs.MemStore{}
	} else {
		panic("unknown driver: " + driver)
	}
	bbs = tagbbs.NewBBS(store)
}

func who(api, user string, params url.Values) (interface{}, error) {
	return user, nil
}

func list(api, user string, params url.Values) (interface{}, error) {
	ids, err := bbs.Query(params.Get("tag"))
	return ids, err
}

func get(api, user string, params url.Values) (interface{}, error) {
	key := params.Get("key")
	post, err := bbs.Get(key, user)
	return post, err
}

func api(handler apiHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS
		if r.Method == "OPTIONS" {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST")
			w.Header().Add("Access-Control-Max-Age", "86400")
			return
		}
		// POST only
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "Method not allowed")
			return
		}

		r.ParseForm()
		log.Println(r.URL.Path, r.Form)

		var (
			result interface{}
			err    error
		)
		switch r.URL.Path {
		case "/login":
			user, pass := r.Form.Get("user"), r.Form.Get("pass")
			randbits := make([]byte, 16)
			_, err = rand.Read(randbits)
			if err != nil {
				panic(err)
			}
			if bbs.Auth(user, pass) {
				sid := hex.EncodeToString(randbits)
				sessionsMutex.Lock()
				sessions[sid] = user
				sessionsMutex.Unlock()
				result, err = sid, nil
			} else {
				result, err = nil, ErrAuthFailed
			}
		case "/logout":
			sessionsMutex.Lock()
			// delete session id
			sid := r.Form.Get("session")
			_, ok := sessions[sid]
			if ok {
				delete(sessions, sid)
				result, err = nil, nil
			} else {
				result, err = nil, ErrUnauthorized
			}
			sessionsMutex.Unlock()
		default:
			sessionsMutex.Lock()
			sid := r.Form.Get("session")
			user, ok := sessions[sid]
			sessionsMutex.Unlock()
			if ok {
				result, err = handler(r.URL.Path, user, r.Form)
			} else {
				result, err = nil, ErrUnauthorized
			}
		}
		encoder := json.NewEncoder(w)
		if err != nil {
			encoder.Encode(M{"error": err.Error()})
		} else {
			encoder.Encode(M{"result": result})
		}
	}
}

func main() {
	flag.Parse()

	log.Println("initializing BBS")
	bbsinit()

	log.Println("listening on " + *flagListen)
	http.HandleFunc("/login", api(nil))
	http.HandleFunc("/logout", api(nil))
	http.HandleFunc("/who", api(who))
	http.HandleFunc("/list", api(list))
	http.HandleFunc("/get", api(get))
	panic(http.ListenAndServe(*flagListen, nil))
}
