package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	_ "github.com/go-sql-driver/mysql"
	"github.com/thinxer/tagbbs"
)

var (
	flagDB = flag.String("db", "mysql://bbs:bbs@/bbs?parseTime=true", "connection string")
	bbs    *tagbbs.BBS
)

var (
	ErrDiscarded = errors.New("Discarded")
)

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

func userauth(user string, password string) bool {
	return bbs.Auth(user, password)
}

func userpubkey(user string, algo string, pubkey []byte) bool {
	type Profile struct {
		AuthorizedKeys string `yaml:"authorized_keys"`
	}
	post, err := bbs.Get("user:"+user, user)
	if err != nil {
		return false
	}
	profile := Profile{}
	err = post.UnmarshalTo(&profile)
	if err != nil {
		return false
	}
	keys := []byte(profile.AuthorizedKeys)
	for {
		var (
			pkey ssh.PublicKey
			ok   bool
		)
		pkey, _, _, keys, ok = ssh.ParseAuthorizedKey(keys)
		if !ok {
			return false
		}
		if pkey.PublicKeyAlgo() == algo && bytes.Compare(ssh.MarshalPublicKey(pkey), pubkey) == 0 {
			return true
		}
	}
}

// Replace with custom terminal implementation.
func usermain(user string, ch ssh.Channel) {
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

	// real logic here
	for {
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
		case "register":
			err := bbs.NewUser(user)
			term.PerrorIf(err)
		case "passwd":
			pass1, _ := term.ReadPassword("New Password:")
			pass2, _ := term.ReadPassword("New Password Again:")
			if pass1 != pass2 {
				term.Perror("Password Mismatch")
			} else {
				term.PerrorIf(bbs.SetUserPass(user, pass1))
			}
		case "who":
			term.Println(user)
			term.PerrorIf(nil)
		case "list":
			name := ""
			if len(cmds) > 1 {
				name = cmds[1]
			}
			ids, err := bbs.Query(name)
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
