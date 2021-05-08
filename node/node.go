package node

import (
	"sync"
)

type NodeState string
type NodeAction string
type NodeStatus string

const (
	NodeStatusInit      NodeStatus = "init"
	NodeStatusInstalled NodeStatus = "installed"
	NodeStatusReady     NodeStatus = "ready"
	NodeStatusUnhealth  NodeStatus = "unhealth"

	NodeStateEnable  NodeState = "enable"
	NodeStateDisable NodeState = "disable"

	NodeActionAdd     NodeAction = "add"
	NodeActionRemove  NodeAction = "remove"
	NodeActionReboot  NodeAction = "reboot"
	NodeActionEnable  NodeAction = "enable"
	NodeActionDisable NodeAction = "disable"
)

var Node_db = make(map[string]*Node)

type Node struct {
	Name       string       `json:"name"`
	UserName   string       `json:"user"`
	Passwd     string       `json:"passwd"`
	Role       string       `json:"role"`
	IpAddress  string       `json:"address"`
	CPU        string       `json:"cpu,omitempty"`
	Memory     string       `json:"memory,omitempty"`
	Disk       string       `json:"disk,omitempty"`
	Status     NodeStatus   `json:"status,omitempty"`
	State      NodeState    `json:"state,omitempty"`
	stateMutex sync.RWMutex `json:"-"`
}

type NodeRequest struct {
	Name      string     `json:"name" form:"name"`
	User      string     `json:"user,omitempty" form:"user,omitempty"`
	Passwd    string     `json:"password,omitempty" form:"password,omitempty"`
	IpAddress string     `json:"ip,omitempty" form:"ip,omitempty"`
	Role      string     `json:"role,omitempty" form:"role,omitempty"`
	Action    NodeAction `json:"action,omitempty" form:"action,omitempty"`
}

// Add a new node
// Args:
//   nodeRequest
// Return:
//   new node pointer
func NewNode(nodeRequest NodeRequest) *Node {

	newNode := Node{
		Name:      nodeRequest.Name,
		IpAddress: nodeRequest.IpAddress,
		UserName:  nodeRequest.User,
		Passwd:    nodeRequest.Passwd,
		Role:      nodeRequest.Role,
		Status:    NodeStatusInit,
		State:     NodeStateEnable,
	}

	return &newNode
}

//Get node pointer by name
//Return nil if not existed
func GetNodeByName(nodeName string) *Node {

	myNode, exists := Node_db[nodeName]
	if exists == false {
		return nil
	} else {
		return myNode
	}

}

//Set node state(enable/disbale)
func (myNode *Node) SetState(state NodeState) {

	myNode.stateMutex.Lock()
	defer myNode.stateMutex.Unlock()

	myNode.State = state

}

//Get node state(enable/disable)
func (myNode *Node) GetState() NodeState {

	myNode.stateMutex.RLock()
	defer myNode.stateMutex.RUnlock()

	return myNode.State

}

//Reboot node
//Return nil if ok, otherwise error
func (myNode *Node) RebootNode() error {
	//TODO
	return nil
}
