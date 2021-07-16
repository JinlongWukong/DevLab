package terminal

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/JinlongWukong/DevLab/utils"
	"github.com/gorilla/websocket"
)

func NewContainerTerminal(ip, name string, port uint16) ContainerTerminal {
	terminal := ContainerTerminal{
		Name:      name,
		IpAddress: ip,
		Port:      port,
	}
	return terminal
}

func (ct *ContainerTerminal) Create() error {
	url := fmt.Sprintf("http://%s:%v/containers/%s/exec", ct.IpAddress, ct.Port, ct.Name)
	payload := "{\"Tty\": true, \"Cmd\": [\"/bin/sh\"], \"AttachStdin\": true, \"AttachStderr\": true, \"AttachStdout\": true}"
	err, resp := utils.HttpSendJsonData(url, "POST", []byte(payload))
	if err != nil {
		log.Println(err)
		return err
	}
	type Container_id struct {
		Id string
	}
	id := &Container_id{}
	err = json.Unmarshal(resp, id)
	if err != nil {
		log.Println(err)
		return err
	}
	ct.Id = id.Id
	return nil
}

func (ct *ContainerTerminal) Start(ws *websocket.Conn) {
	dockerAddr := fmt.Sprintf("%s:%v", ct.IpAddress, ct.Port)
	conn, err := net.Dial("tcp", dockerAddr)
	if err != nil {
		log.Printf("tcp connect to docker api failed: %v", err.Error())
		return
	}
	data := "{\"Tty\":true}"
	_, err = conn.Write([]byte(fmt.Sprintf("POST /exec/%s/start HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nContent-Length: %s\r\n\r\n%s",
		ct.Id, dockerAddr, fmt.Sprint(len([]byte(data))), data)))
	if err != nil {
		log.Printf("failed to start an exec instance: %v", err.Error())
		return
	}

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	defer func() {
		conn.Close()
		ws.Close()
	}()

	// websocket -> tcp
	go func() {
		for {
			_, p, err := ws.ReadMessage()
			if err != nil {
				return
			}
			_, err = conn.Write(p)
			if err != nil {
				return
			}
		}
	}()

	// tcp -> websocket
	for {
		b := make([]byte, 512)
		conn.Read(b)
		ws.WriteMessage(websocket.TextMessage, b)
	}

}
