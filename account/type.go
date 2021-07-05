package account

import (
	"sync"

	"github.com/JinlongWukong/DevLab/k8s"
	"github.com/JinlongWukong/DevLab/saas"
	"github.com/JinlongWukong/DevLab/vm"
)

type Account struct {
	Name                string               `json:"name"`
	OneTimePass         string               `json:"-"`
	Role                string               `json:"role"`
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
	Name string `form:"name" json:"name" binding:"required"`
	Role string `form:"role" json:"role" binding:"required"`
}
