package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	flagDB     = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")
	flagListen = flag.String("listen", ":8023", "address to listen on")
	flagCert   = flag.String("cert", "", "HTTPS: Certificate")
	flagKey    = flag.String("key", "", "HTTPS: Key")
)

var (
	ErrUnauthorized = errors.New("Unauthorized")
)

var (
	sessions      = make(map[string]string)
	sessionsMutex sync.Mutex
)

type V []interface{}
type M map[string]interface{}
type apiHandler func(api, user string, params url.Values) (interface{}, error)

func api(handler apiHandler, auth bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		remoteIp := r.Header.Get("X-Real-IP")
		if len(remoteIp) == 0 {
			remoteIp = r.RemoteAddr
		}
		log.Println(remoteIp, r.URL.Path, r.ContentLength)
		// Allow CORS
		w.Header().Add("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
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
		r.ParseForm()

		var (
			result interface{}
			err    error
		)
		if auth {
			sessionsMutex.Lock()
			sid := r.Form.Get("session")
			user, ok := sessions[sid]
			sessionsMutex.Unlock()
			if ok {
				result, err = handler(r.URL.Path, user, r.Form)
			} else {
				result, err = nil, ErrUnauthorized
			}
		} else {
			result, err = handler(r.URL.Path, "", r.Form)
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

	http.HandleFunc("/login", api(login, false))
	http.HandleFunc("/logout", api(logout, false))
	http.HandleFunc("/version", api(version, false))
	http.HandleFunc("/register", api(register, false))
	http.HandleFunc("/passwd", api(passwd, true))
	http.HandleFunc("/who", api(who, true))
	http.HandleFunc("/list", api(list, true))
	http.HandleFunc("/get", api(get, true))
	http.HandleFunc("/put", api(put, true))

	if len(*flagCert) > 0 {
		log.Println("listening on " + *flagListen + " with HTTPS")
		panic(http.ListenAndServeTLS(*flagListen, *flagCert, *flagKey, nil))
	} else {
		log.Println("listening on " + *flagListen + " without HTTPS")
		panic(http.ListenAndServe(*flagListen, nil))
	}
}
