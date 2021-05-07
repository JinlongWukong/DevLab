package node

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/JinlongWukong/CloudLab/utils"
)

var Node_db = make(map[string]*Node)

type Node struct {
	Name      string `json:"name,omitempty"`
	CPU       int    `json:"cpu,omitempty"`
	Memory    int    `json:"memory,omitempty"`
	Disk      int    `json:"disk,omitempty"`
	IpAddress string `json:"ip_address,omitempty"`
	Status    string `json:"status,omitempty"`
	UserName  string `json:"user_name,omitempty"`
	Passwd    string `json:"passwd,omitempty"`
	Role      string `json:"role,omitempty"`
}

type NodeRequest struct {
	Name      string `json:"name,omitempty"`
	User      string `json:"user,omitempty"`
	Passwd    string `json:"password,omitempty"`
	IpAddress string `json:"ip,omitempty"`
	Role      string `json:"role,omitempty"`
	Action    string `json:"action,omitempty"`
	Status    string `json:"status,omitempty"`
}

type NodeRequestGetNode struct {
	Name string `form:"name"`
}

// Add a new node
// Args:
//   nodeRequest
// Return:
//   node pointer
//   error
func NewNode(nodeRequest NodeRequest) (*Node, error) {

	log.Printf("Start install node %v", nodeRequest.Name)

	payload, _ := json.Marshal(map[string]interface{}{
		"Ip":     nodeRequest.IpAddress,
		"Pass":   nodeRequest.Passwd,
		"User":   nodeRequest.User,
		"Role":   nodeRequest.Role,
		"Action": "install",
	})

	log.Println("Remote http call to install node")
	//Always create node struct, set status accordingly
	newNode := Node{Name: nodeRequest.Name, IpAddress: nodeRequest.IpAddress, UserName: nodeRequest.User,
		Passwd: nodeRequest.Passwd, Role: nodeRequest.Role}

	err := utils.HttpSendJsonData("http://10.124.44.167:9134/host", "POST", payload)
	if err != nil {
		log.Println(err)
		newNode.Status = fmt.Sprint(err)
		return &newNode, err
	} else {
		newNode.Status = "running"
		return &newNode, nil
	}

}

func GetNodeByName(nodeName string) *Node {

	//return Node{IpAddress: "127.0.0.1", UserName: "root", Passwd: "Cisco123!", Role: "compute"}
	_, exists := Node_db[nodeName]
	if exists == false {
		return nil
	} else {
		return Node_db[nodeName]
	}

}

func (myNode Node) RebootNode() error {
	//TODO
	return nil
}
