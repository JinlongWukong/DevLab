package node

import "sync"

type NodeState string
type NodeAction string
type NodeStatus string

const (
	NodeStatusInit      NodeStatus = "init"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusInstalled NodeStatus = "installed"
	NodeStatusReady     NodeStatus = "ready"
	NodeStatusUnhealth  NodeStatus = "unhealth"
	NodeStatusOverload  NodeStatus = "overload"

	NodeStateEnable  NodeState = "enable"
	NodeStateDisable NodeState = "disable"

	NodeActionAdd     NodeAction = "add"
	NodeActionRemove  NodeAction = "remove"
	NodeActionReboot  NodeAction = "reboot"
	NodeActionEnable  NodeAction = "enable"
	NodeActionDisable NodeAction = "disable"
	NodePortRangeMin             = 20000
	NodePortRangeMax             = 25000
)

type Node struct {
	Name        string         `json:"name"`
	UserName    string         `json:"user"`
	Passwd      string         `json:"passwd"`
	Role        string         `json:"role"`
	IpAddress   string         `json:"address"`
	OSType      string         `json:"os"`
	Subnet      string         `json:"subnet"`
	CPU         int32          `json:"cpu"`
	Memory      int32          `json:"memory"`
	Disk        int32          `json:"disk"`
	CpuUsed     int32          `json:"cpuUsed"`
	MemUsed     int32          `json:"memUsed"`
	DiskUsed    int32          `json:"diskUsed"`
	PortMap     map[int]string `json:"portMap"`
	Status      NodeStatus     `json:"status"`
	State       NodeState      `json:"state"`
	statusMutex sync.RWMutex   `json:"-"`
	stateMutex  sync.RWMutex   `json:"-"`
	portMutex   sync.Mutex     `json:"-"`
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

type NodeUsage struct {
	CpuLoad   float64 `json:"cpu_load"`
	MemAvail  int     `json:"memory_avail"`
	DiskUsage string  `json:"disk_usage"`
}
