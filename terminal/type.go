package terminal

import "golang.org/x/crypto/ssh"

type ptyRequestParams struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

type SSHTerminal struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	IpAddress string `json:"ipaddress"`
	Port      uint16 `json:"port"`
	Session   *ssh.Session
	Client    *ssh.Client
	channel   ssh.Channel
}

type ContainerTerminal struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	IpAddress string `json:"ipaddress"`
	Port      uint16 `json:"port"`
}
