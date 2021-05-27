package account

import (
	"fmt"
	"log"
	"sync"

	"github.com/JinlongWukong/CloudLab/notification"
	"github.com/JinlongWukong/CloudLab/vm"
)

var AccountDB = AccountMap{Map: make(map[string]*Account)}

type AccountMap struct {
	Map  map[string]*Account `json:"account"`
	lock sync.RWMutex        `json:"-"`
}

type AccountMapItem struct {
	Key   string
	Value *Account
}

func (m *AccountMap) Set(key string, value *Account) {

	m.lock.Lock()
	defer m.lock.Unlock()

	m.Map[key] = value

}

func (m *AccountMap) Get(key string) (account *Account, exists bool) {

	m.lock.RLock()
	defer m.lock.RUnlock()

	account, exists = m.Map[key]
	return

}

func (m *AccountMap) Del(key string) {

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.Map, key)

}

// Iter iterates over the items in a concurrent map
// Each item is sent over a channel, so that
// we can iterate over the map using the builtin range keyword
func (m *AccountMap) Iter() <-chan AccountMapItem {
	c := make(chan AccountMapItem)

	f := func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		for k, v := range m.Map {
			c <- AccountMapItem{k, v}
		}
		close(c)
	}
	go f()

	return c
}

type Account struct {
	Name          string               `json:"name"`
	Role          string               `json:"role"`
	VM            []*vm.VirtualMachine `json:"vm"`
	lockerVMSlice sync.Mutex           `json:"-"`
	StatusVm      string               `json:"-"`
}

func (a *Account) GetNumbersOfVm() int {

	a.lockerVMSlice.Lock()
	defer a.lockerVMSlice.Unlock()

	return len(a.VM)
}

func (a *Account) GetVmNameList() []string {

	a.lockerVMSlice.Lock()
	defer a.lockerVMSlice.Unlock()

	vmNames := make([]string, 0)
	for _, v := range a.VM {
		vmNames = append(vmNames, v.Name)
	}

	return vmNames
}

func (a *Account) GetVmByName(name string) (*vm.VirtualMachine, error) {

	a.lockerVMSlice.Lock()
	defer a.lockerVMSlice.Unlock()

	for _, v := range a.VM {
		if v.Name == name {
			return v, nil
		}
	}

	return nil, fmt.Errorf("VM %v not found", name)
}

func (a *Account) AppendVM(vm *vm.VirtualMachine) {

	a.lockerVMSlice.Lock()
	defer a.lockerVMSlice.Unlock()

	a.VM = append(a.VM, vm)
}

func (a *Account) RemoveVmByName(name string) error {

	a.lockerVMSlice.Lock()
	defer a.lockerVMSlice.Unlock()

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

func (a *Account) Iter() <-chan *vm.VirtualMachine {
	c := make(chan *vm.VirtualMachine)

	f := func() {
		a.lockerVMSlice.Lock()
		defer a.lockerVMSlice.Unlock()

		for _, v := range a.VM {
			c <- v
		}
		close(c)
	}
	go f()

	return c
}

func (a *Account) SendNotification(msg string) {

	notification.SendNotification(notification.Message{Target: a.Name + "@cisco.com", Text: msg})

}
