package account

import (
	"fmt"

	"github.com/JinlongWukong/CloudLab/vm"
)

type Account struct {
	Name     string
	Role     string
	VM       []*vm.VirtualMachine
	StatusVm string
}

func (a Account) GetNumbersOfVm() int {

	return len(a.VM)
}

func (a Account) GetVmNameList() []string {

	vmNames := make([]string, 0)
	for _, v := range a.VM {
		vmNames = append(vmNames, v.Name)
	}

	return vmNames
}

func (a Account) GetVmByName(name string) (vm.VirtualMachine, error) {

	for _, v := range a.VM {
		if v.Name == name {
			return *v, nil
		}
	}

	return vm.VirtualMachine{}, fmt.Errorf("VM %v not found", name)
}
