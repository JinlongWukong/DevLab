package node

import "sync"

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

type Node struct {
	Name       string       `json:"name"`
	UserName   string       `json:"user"`
	Passwd     string       `json:"passwd"`
	Role       string       `json:"role"`
	IpAddress  string       `json:"address"`
	OSType     string       `json:"os"`
	CPU        int32        `json:"cpu"`
	Memory     int32        `json:"memory"`
	Disk       int32        `json:"disk"`
	CpuUsed    int32        `json:"cpuUsed"`
	MemUsed    int32        `json:"memUsed"`
	DiskUsed   int32        `json:"diskUsed"`
	Status     NodeStatus   `json:"status"`
	State      NodeState    `json:"state"`
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

type NodeInfo struct {
	CPU    int32  `json:"cpu"`
	Memory int32  `json:"memory"`
	Disk   int32  `json:"disk"`
	OSType string `json:"type"`
}
