package k8s

import (
	"sync"
	"time"
)

type K8sStatus string

const (
	K8sStatusInit          K8sStatus = "init"
	K8sStatusBootingVm     K8sStatus = "bootingvm"
	K8sStatusBootVmFailed  K8sStatus = "bootvmFailed"
	K8sStatusInstalling    K8sStatus = "installing"
	K8sStatusInstallFailed K8sStatus = "installFailed"
	K8sStatusRunning       K8sStatus = "running"
	K8sStatusDeleting      K8sStatus = "deleting"
)

type K8S struct {
	Name             string        `json:"name"`
	Version          string        `json:"version"`
	NumOfContronller uint16        `json:"numOfContronller"`
	NumOfWorker      uint16        `json:"numOfWorker"`
	Lifetime         time.Duration `json:"lifeTime"`
	Status           K8sStatus     `json:"status"`
	HostVm           string        `json:"hostVm"`
	sync.RWMutex     `json:"-"`
}

type K8sRequest struct {
	Account          string `form:"account" json:"account" binding:"required"`
	Version          string `form:"version" json:"version" binding:"required"`
	NumOfContronller uint16 `form:"numOfContronller" json:"numOfContronller" binding:"omitempty,max=5"`
	NumOfWorker      uint16 `form:"numOfWorker" json:"numOfWorker" binding:"omitempty,max=100"`
	Duration         int    `form:"duration" json:"duration" binding:"omitempty"`
}

type K8sRequestAction struct {
	Account string `form:"account" json:"account" binding:"required"`
	Name    string `form:"name" json:"name"`
	Action  string `form:"action,omitempty" json:"action,omitempty"`
}
