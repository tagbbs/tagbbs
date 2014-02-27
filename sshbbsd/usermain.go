package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"reflect"
	"strings"
	"github.com/tagbbs/tagbbs/rkv"

	"code.google.com/p/go.crypto/ssh"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tagbbs/tagbbs"
	"github.com/tagbbs/tagbbs/auth"
)

const metaUser = "."

var (
	flagDB = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")
	bbs    *tagbbs.BBS
	auths  auth.AuthenticationList
)

var (
	ErrDiscarded = errors.New("Discarded")
)

func bbsinit() {
	bbs = tagbbs.NewBBSFromString(*flagDB)
	auths = auth.AuthenticationList{
		auth.Password{rkv.ScopedStore{bbs.Storage, "userpass:"}},
		ProfilePublicKeyAuth{bbs},
	}
}

// Replace with custom terminal implementation.
func usermain(user, remoteAddr string, ch ssh.Channel) {
	// catch all errors so that the main server won't crash
	defer func() {
		if err := recover(); err != nil {
			log.Println("!!! Err:", err)
		}
	}()

	// create a term on top of the connection
	term := NewTerminal(ch, "> ")
	serverTerm := &ssh.ServerTerminal{
		Term:    term,
		Channel: ch,
	}

	// ready to accept the connection
	ch.Accept()
	defer ch.Close()

	sid, err := bbs.SessionManager.Request(tagbbs.Session{
		User:       user,
		UserAgent:  "Terminal/SSH",
		Capability: tagbbs.CapRead | tagbbs.CapPost,
		RemoteAddr: remoteAddr,
	})
	if err != nil {
		term.PerrorIf(err)
		return
	}
	defer func() {
		bbs.SessionManager.Revoke(sid)
	}()

	// real logic here
	name, version, err := bbs.Version()
	if err != nil {
		panic(err)
	} else {
		term.Printf("%s: %s\r\n", name, version)
	}

	// meta user
	if user == metaUser {
		term.Println("Registering...")
		term.SetPrompt("Username: ")
		newuser, err := serverTerm.ReadLine()
		if err != nil {
			return
		}
		pass1, err := term.ReadPassword("Password: ")
		if err != nil {
			return
		}
		pass2, err := term.ReadPassword("Retype Password: ")
		if err != nil {
			return
		}
		if pass1 != pass2 {
			term.Perror("Password Mismatch")
		}
		pw := auths.Of(reflect.TypeOf(auth.Password{})).(auth.Password)
		term.PerrorIf(pw.New(newuser, pass1))
		return
	}

	for {
		// check if session still valid
		_, err := bbs.SessionManager.Get(sid)
		if err != nil {
			term.Perror(err)
			break
		}

		// REPL
		line, err := serverTerm.ReadLine()
		if err == io.EOF {
			return
		} else if err != nil {
			log.Println("readLine err:", err)
			continue
		}
		// deal with the command
		cmds := strings.Split(line, " ")
		switch cmds[0] {
		case "help":
			term.Println("Available Commands:")
			term.Println("help\r\nregister\r\npasswd\r\nwho\r\nlist [tag]\r\nget key\r\nput [key]")
		case "passwd":
			pass1, _ := term.ReadPassword("New Password:")
			pass2, _ := term.ReadPassword("New Password Again:")
			if pass1 != pass2 {
				term.Perror("Password Mismatch")
			} else {
				pw := auths.Of(reflect.TypeOf(auth.Password{})).(auth.Password)
				term.PerrorIf(pw.Set(user, pass1))
			}
		case "who":
			term.Println(user)
			term.PerrorIf(nil)
		case "list":
			query := ""
			if len(cmds) > 1 {
				query = strings.Join(cmds[1:], " ")
			}
			ids, _, err := bbs.Query(query)
			for _, id := range ids {
				p, _ := bbs.Get(id, user)
				fm := p.FrontMatter()
				if fm == nil {
					continue
				}
				paddedTitle := fm.Title
				width := StringWidth(paddedTitle)
				if width < 40 {
					paddedTitle += strings.Repeat(" ", 40-width)
				}
				term.Printf("%-12s %s %v %v\r\n", id, paddedTitle, fm.Authors, fm.Tags)
			}
			term.PerrorIf(err)
		case "get":
			key := ""
			if len(cmds) > 1 {
				key = cmds[1]
			}
			post, err := bbs.Get(key, user)

			if err != nil {
				term.Perror(err)
			} else {
				term.Pokay(post.Rev, post.Timestamp)
				term.WriteUnix(post.Content)
			}
		case "put":
			var (
				key  string
				post tagbbs.Post
			)
			if len(cmds) > 1 {
				key = cmds[1]
			}
			if len(key) > 0 {
				var err error
				post, err = bbs.Get(key, user)
				if err != nil {
					term.Perror(err)
				}
			}
			term.SetPrompt("")
			if post.Rev > 0 {
				term.Pokay("Editing ", key, ", rev: ", post.Rev)
			}
			term.Pokay("Type EOF to end the file. Type DISCARD to cancel.")
			content, err := readutil(serverTerm, "EOF")
			if err == nil {
				post.Rev++
				post.Content = []byte(content)
				if len(key) == 0 {
					key = bbs.NewPostKey()
				}
				term.PerrorIf(bbs.Put(key, post, user))
			} else {
				term.Perror(err)
			}
			term.SetPrompt("> ")
		}
	}
}

func readutil(term *ssh.ServerTerminal, until string) (content string, err error) {
	var line string
	for {
		line, err = term.ReadLine()
		if err != nil {
			return
		}
		if line == "EOF" {
			break
		} else if line == "DISCARD" {
			content = ""
			err = ErrDiscarded
			break
		}
		content += line + "\n"
	}
	return
}
