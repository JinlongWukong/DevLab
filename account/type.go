package account

import (
	"sync"

	"github.com/JinlongWukong/DevLab/k8s"
	"github.com/JinlongWukong/DevLab/saas"
	"github.com/JinlongWukong/DevLab/vm"
)

type RoleType string

const (
	RoleAdmin RoleType = "admin"
	RoleGuest RoleType = "guest"
)

type Account struct {
	Name                string               `json:"name"`
	OneTimePass         string               `json:"-"`
	Role                RoleType             `json:"role"`
	Contract            string               `json:"contract"`
	VM                  []*vm.VirtualMachine `json:"vm"`
	K8S                 []*k8s.K8S           `json:"k8s"`
	Software            []*saas.Software     `json:"software"`
	lockerVMSlice       sync.Mutex           `json:"-"`
	lockerK8SSlice      sync.Mutex           `json:"-"`
	lockerSoftwareSlice sync.Mutex           `json:"-"`
	sync.Mutex          `json:"-"`
}

type AccountRequest struct {
	Name     string   `form:"name" json:"name" binding:"required"`
	Role     RoleType `form:"role" json:"role" binding:"required"`
	Contract string   `form:"contract,omitempty" json:"contract,omitempty"`
}
