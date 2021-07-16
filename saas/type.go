package saas

import "sync"

type SoftwareStatus string
type SoftwareAction string

const (
	SoftwareStatusInit          SoftwareStatus = "init"
	SoftwareStatusScheduled     SoftwareStatus = "scheduled"
	SoftwareStatusBootingVm     SoftwareStatus = "bootingvm"
	SoftwareStatusBootVmFailed  SoftwareStatus = "bootvmFailed"
	SoftwareStatusInstalling    SoftwareStatus = "installing"
	SoftwareStatusInstallFailed SoftwareStatus = "installFailed"
	SoftwareStatusRunning       SoftwareStatus = "running"
	SoftwareStatusStopped       SoftwareStatus = "stopped"
	SoftwareStatusDeleting      SoftwareStatus = "deleting"
	SoftwareStatusNotFound      SoftwareStatus = "notFound"
	SoftwareStatusError         SoftwareStatus = "error"
	SoftwareStatusUnknown       SoftwareStatus = "unknown"

	SoftwareActionStart   SoftwareAction = "start"
	SoftwareActionStop    SoftwareAction = "stop"
	SoftwareActionRestart SoftwareAction = "restart"
	SoftwareActionDelete  SoftwareAction = "delete"
	SoftwareActionGet     SoftwareAction = "get"
)

type Software struct {
	Name            string            `json:"name"`
	Kind            string            `json:"kind"`
	Backend         string            `json:"backend"`
	Version         string            `json:"version"`
	Address         string            `json:"address"`
	Node            string            `json:"node"`
	CPU             uint8             `json:"cpu"`
	Memory          uint32            `json:"memory"`
	Status          SoftwareStatus    `json:"status"`
	PortMapping     map[string]string `json:"port_mapping"`
	AdditionalInfor map[string]string `json:"additional_infor"`
	statusMutex     sync.RWMutex      `json:"-"`
	sync.Mutex      `json:"-"`
}

type SoftwareRequest struct {
	Kind    string `form:"kind" json:"kind" binding:"required"`
	Version string `form:"version" json:"version" binding:"required"`
	CPU     uint8  `form:"cpu" json:"cpu" binding:"required,min=1,max=20"`
	Memory  uint32 `form:"memory" json:"memory" binding:"required,min=10,max=65536"`
}

type SoftwareRequestAction struct {
	Account string         `form:"account" json:"account"`
	Name    string         `form:"name" json:"name"`
	Action  SoftwareAction `form:"action" json:"action"`
}

/*
{
    "additional_infor": {
        "admin_password": "33c426556ee546949703391ce86a9cb6"
    },
    "address": "172.19.2.3",
    "port_mapping": [
        "50000/tcp -> 0.0.0.0:49187",
        "8080/tcp -> 0.0.0.0:49188"
    ]
}
*/
type SoftwareInfo struct {
	Status          string            `json:"status"`
	Address         string            `json:"address"`
	PortMapping     []string          `json:"port_mapping"`
	AdditionalInfor map[string]string `json:"additional_infor,omitempty"`
}
