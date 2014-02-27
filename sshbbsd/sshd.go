package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/url"

	"code.google.com/p/go.crypto/ssh"
)

var (
	flagListen = flag.String("listen", ":1481", "Address to listen on")
	flagKey    = flag.String("key", "id_rsa", "Private Key for SSH Server")
)

func main() {
	flag.Parse()

	log.Println("initializing bbs")
	bbsinit()

	log.Println("initializing sshd")
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn *ssh.ServerConn, user, algo string, pubkey []byte) bool {
			params := url.Values{}
			params.Set("user", user)
			params.Set("algo", algo)
			params.Set("pubkey", string(pubkey))
			_, err := auths.Auth(params)
			return err == nil
		},
		KeyboardInteractiveCallback: func(conn *ssh.ServerConn, user string, client ssh.ClientKeyboardInteractive) bool {
			if user == metaUser {
				return true
			}
			name, _, err := bbs.Version()
			if err != nil {
				return false
			}
			ans, err := client.Challenge(
				user,
				`Welcome to `+name+`!
For registering, please re-login with . (a single dot) as username.
You're trying to login as `+user+".",
				[]string{"Password:"},
				[]bool{false},
			)
			if err != nil {
				log.Println(err)
				return false
			}
			params := url.Values{}
			params.Set("user", user)
			params.Set("pass", ans[0])
			_, err = auths.Auth(params)
			return err == nil
		},
	}

	// Load server certificate.
	pemBytes, err := ioutil.ReadFile(*flagKey)
	if err != nil {
		log.Fatal("Failed to load private key:", err)
	}
	if err = config.SetRSAPrivateKey(pemBytes); err != nil {
		log.Fatal("Failed to parse private key:", err)
	}

	// Listen
	log.Println("listening on", *flagListen)
	conn, err := ssh.Listen("tcp", *flagListen, config)
	if err != nil {
		log.Fatal("Failed to listen for connection:", err)
	}
	for {
		sConn, err := conn.Accept()
		if err != nil {
			log.Println("Failed to accept incoming connection:", err)
			continue
		}

		go handleServerConn(sConn)
	}
}

func handleServerConn(sConn *ssh.ServerConn) {
	defer sConn.Close()

	log.Println("from:", sConn.RemoteAddr())
	if err := sConn.Handshake(); err != nil {
		log.Println("failed to handshake", err)
		return
	}

	for {
		ch, err := sConn.Accept()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("handleServerConn Accept:", err)
			break
		}

		// We need a "session" type for shell
		if ch.ChannelType() != "session" {
			ch.Reject(ssh.UnknownChannelType, "unknown channel type")
			break
		}

		go usermain(sConn.User, sConn.RemoteAddr().String(), ch)
	}
}
