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
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/thinxer/tagbbs"
)

var (
	flagDB     = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")
	flagListen = flag.String("listen", ":8023", "address to listen on")
	flagCert   = flag.String("cert", "", "HTTPS: Certificate")
	flagKey    = flag.String("key", "", "HTTPS: Key")

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
	store, err := tagbbs.NewStore(*flagDB)
	if err != nil {
		panic(err)
	}
	bbs = tagbbs.NewBBS(store)
	log.Println(bbs.Version())
}

func who(api, user string, params url.Values) (interface{}, error) {
	return user, nil
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

func api(handler apiHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS
		if r.Method == "OPTIONS" {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			w.Header().Add("Access-Control-Max-Age", "86400")
			return
		}
		// POST only
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "Method not allowed")
			return
		}

		w.Header().Add("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		log.Println(r.URL.Path, r.Form)

		var (
			result interface{}
			err    error
		)
		switch r.URL.Path {
		case "/version":
			var name, ver string
			name, ver, err = bbs.Version()
			result = M{
				"name":    name,
				"version": ver,
			}
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

	http.HandleFunc("/login", api(nil))
	http.HandleFunc("/logout", api(nil))
	http.HandleFunc("/version", api(nil))
	http.HandleFunc("/who", api(who))
	http.HandleFunc("/list", api(list))
	http.HandleFunc("/get", api(get))
	http.HandleFunc("/put", api(put))

	if len(*flagCert) > 0 {
		log.Println("listening on " + *flagListen + " with HTTPS")
		panic(http.ListenAndServeTLS(*flagListen, *flagCert, *flagKey, nil))
	} else {
		log.Println("listening on " + *flagListen + " without HTTPS")
		panic(http.ListenAndServe(*flagListen, nil))
	}
}
