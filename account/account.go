package account

import (
	"fmt"
	"log"

	"github.com/JinlongWukong/CloudLab/vm"
)

var Account_db = make(map[string]*Account)

type Account struct {
	Name     string
	Role     string
	VM       []*vm.VirtualMachine
	StatusVm string `json:"-"`
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

func (a Account) GetVmByName(name string) (*vm.VirtualMachine, error) {

	log.Println(a.VM)
	for _, v := range a.VM {
		if v.Name == name {
			return v, nil
		}
	}

	return nil, fmt.Errorf("VM %v not found", name)
}

func (a *Account) RemoveVmByName(name string) error {

	//To remove item from slice, this is a fast version (changes order)
	for i, v := range a.VM {
		if v.Name == name {
			// Remove the element at index i from a.
			a.VM[i] = a.VM[len(a.VM)-1] // Copy last element to index i.
			a.VM[len(a.VM)-1] = nil     // Erase last element (write zero value).
			a.VM = a.VM[:len(a.VM)-1]   // Truncate slice.
			log.Printf("VM %v has been removed from account %v", name, a.Name)
			return nil
		}
	}

	return fmt.Errorf("VM %v not found", name)
}
