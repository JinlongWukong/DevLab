package terminal

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

// Wrapper of ws, to implement io.Writer
type wapperWS struct {
	ws *websocket.Conn
}

func (self *wapperWS) Write(buf []byte) (int, error) {
	err := self.ws.WriteMessage(websocket.TextMessage, buf)
	return len(buf), err
}

func NewSSHTerminal(username, password, ip string, port uint16) SSHTerminal {
	terminal := SSHTerminal{}
	terminal.Username = username
	terminal.Password = password
	terminal.IpAddress = ip
	terminal.Port = port
	return terminal
}

// Dial ssh connection
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

// Setup default interactive shell terminal
// ssh user@ip
func (st *SSHTerminal) NewShellTerminal() error {
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

// Setup interactive cmd terminal
// ssh -t user@ip  <interactive cmd>
func (st *SSHTerminal) NewInteractiveCmdTerminal(ws *websocket.Conn, cmd string) {
	defer ws.Close()
	session, err := st.Client.NewSession()
	if err != nil {
		log.Println(err)
		return
	}
	defer session.Close()
	st.Session = session

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	targetStdout, _ := session.StdoutPipe()
	targetStdin, _ := session.StdinPipe()

	// ssh stdout => websocket
	go io.Copy(&wapperWS{ws}, targetStdout)
	/*
		go func() {
			for {
				b := make([]byte, 1024)
				targetStdout.Read(b)
				if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
					log.Println(err)
					return
				}
			}
		}()
	*/

	// websocket => ssh stdin
	go func() {
		for {
			_, p, err := ws.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			targetStdin.Write(p)
		}
	}()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", 150*8, 30*8, modes); err != nil {
		log.Printf("request for pseudo terminal failed: %s", err)
		return
	}

	if err := session.Run(cmd); err != nil {
		log.Printf("remote exec interactive cmd %v failed: %s", cmd, err)
		return
	}

	log.Printf("remote exec interactive cmd %v exited normally", cmd)
}

// Run remote cmd terminal
// ssh user@ip  <cmd>
func (st *SSHTerminal) NewCmdTerminal(ws *websocket.Conn, cmd string) error {
	defer ws.Close()
	session, err := st.Client.NewSession()
	if err != nil {
		log.Println(err)
		return err
	}
	defer session.Close()
	st.Session = session

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	targetStdout, _ := session.StdoutPipe()

	// ssh stdout => websocket
	go io.Copy(&wapperWS{ws}, targetStdout)

	if err := session.Run(cmd); err != nil {
		log.Printf("remote exec cmd %v failed: %s", cmd, err)
		return err
	}
	// sleep a while for stdout to ws
	time.Sleep(1 * time.Second)
	log.Printf("remote exec cmd %v exited normally", cmd)
	return nil
}

// websocket <=> ssh channel, as the next step of NewShellTerminal
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
