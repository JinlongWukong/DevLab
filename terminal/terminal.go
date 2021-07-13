package terminal

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

func NewSSHTerminal(username, password, ip string, port uint16) SSHTerminal {
	terminal := SSHTerminal{}
	terminal.Username = username
	terminal.Password = password
	terminal.IpAddress = ip
	terminal.Port = port
	return terminal
}

func (st *SSHTerminal) Connect() error {
	authM := make([]ssh.AuthMethod, 0)
	authM = append(authM, ssh.Password(st.Password))
	clientConfig := &ssh.ClientConfig{
		User:    st.Username,
		Auth:    authM,
		Timeout: 5 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	addr := fmt.Sprintf("%s:%d", st.IpAddress, st.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return err
	} else {
		st.Client = client
	}

	return nil
}

func (st *SSHTerminal) NewTerminal() error {
	session, err := st.Client.NewSession()
	if err != nil {
		log.Println(err)
		return nil
	}
	st.Session = session
	channel, inRequests, err := st.Client.OpenChannel("session", nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	st.channel = channel
	go func() {
		for req := range inRequests {
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	var modeList []byte
	for k, v := range modes {
		kv := struct {
			Key byte
			Val uint32
		}{k, v}
		modeList = append(modeList, ssh.Marshal(&kv)...)
	}
	modeList = append(modeList, 0)

	req := ptyRequestParams{
		Term:     "xterm",
		Columns:  150,
		Rows:     30,
		Width:    uint32(150 * 8),
		Height:   uint32(30 * 8),
		Modelist: string(modeList),
	}
	ok, err := channel.SendRequest("pty-req", true, ssh.Marshal(&req))
	if !ok || err != nil {
		log.Println(err)
		return nil
	}
	ok, err = channel.SendRequest("shell", true, nil)
	if !ok || err != nil {
		log.Println(err)
		return nil
	}

	return nil
}

// Loop control, websocket <> ssh connection
func (st *SSHTerminal) Ws2ssh(ws *websocket.Conn) {

	// get user input
	go func() {
		for {
			_, p, err := ws.ReadMessage()
			if err != nil {
				return
			}
			_, err = st.channel.Write(p)
			if err != nil {
				return
			}
		}
	}()

	// return ssh result to user
	go func() {
		br := bufio.NewReader(st.channel)
		buf := []byte{}
		t := time.NewTimer(time.Microsecond * 100)
		defer t.Stop()

		r := make(chan rune)

		go func() {
			defer st.Client.Close()
			defer st.Session.Close()

			for {
				x, size, err := br.ReadRune()
				if err != nil {
					log.Println(err)
					//ws.WriteMessage(1, []byte("\033[31mConnection closed!\033[0m"))
					ws.Close()
					return
				}
				if size > 0 {
					r <- x
				}
			}
		}()

		// main loop
		for {
			select {
			// if buf lenght not 0, write data into ws every each 100ms
			case <-t.C:
				if len(buf) != 0 {
					err := ws.WriteMessage(websocket.TextMessage, buf)
					buf = []byte{}
					if err != nil {
						log.Println(err)
						return
					}
				}
				t.Reset(time.Microsecond * 100)

			case d := <-r:
				if d != utf8.RuneError {
					p := make([]byte, utf8.RuneLen(d))
					utf8.EncodeRune(p, d)
					buf = append(buf, p...)
				} else {
					buf = append(buf, []byte("@")...)
				}
			}
		}
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
}
