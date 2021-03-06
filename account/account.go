package account

import (
	"fmt"
	"log"

	"github.com/JinlongWukong/DevLab/auth"
	"github.com/JinlongWukong/DevLab/config"
	"github.com/JinlongWukong/DevLab/k8s"
	"github.com/JinlongWukong/DevLab/notification"
	"github.com/JinlongWukong/DevLab/saas"
	"github.com/JinlongWukong/DevLab/vm"
)

// factory to make account
func newAccount(accountRequest AccountRequest) (*Account, error) {
	switch accountRequest.Role {
	case RoleAdmin:
	case RoleGuest:
	default:
		return nil, fmt.Errorf("account role not valid")
	}

	newAccount := &Account{
		Name: accountRequest.Name,
		Role: accountRequest.Role,
	}
	if config.Notification.Kind == "webex" {
		newAccount.Contract = newAccount.Name + "@cisco.com"
	} else {
		newAccount.Contract = accountRequest.Contract
	}
	return newAccount, nil
}

//VM part
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

//K8S part
func (a *Account) GetNumbersOfK8s() int {

	a.lockerK8SSlice.Lock()
	defer a.lockerK8SSlice.Unlock()

	return len(a.K8S)
}

func (a *Account) GetK8sNameList() []string {

	a.lockerK8SSlice.Lock()
	defer a.lockerK8SSlice.Unlock()

	k8sNames := make([]string, 0)
	for _, v := range a.K8S {
		k8sNames = append(k8sNames, v.Name)
	}

	return k8sNames
}

func (a *Account) GetK8sByName(name string) (*k8s.K8S, error) {

	a.lockerK8SSlice.Lock()
	defer a.lockerK8SSlice.Unlock()

	for _, v := range a.K8S {
		if v.Name == name {
			return v, nil
		}
	}

	return nil, fmt.Errorf("K8S %v not found", name)
}

func (a *Account) AppendK8S(k8s *k8s.K8S) {

	a.lockerK8SSlice.Lock()
	defer a.lockerK8SSlice.Unlock()

	a.K8S = append(a.K8S, k8s)
}

func (a *Account) RemoveK8sByName(name string) error {

	a.lockerK8SSlice.Lock()
	defer a.lockerK8SSlice.Unlock()

	//To remove item from slice, this is a fast version (changes order)
	for i, v := range a.K8S {
		if v.Name == name {
			// Remove the element at index i from a.
			a.K8S[i] = a.K8S[len(a.K8S)-1] // Copy last element to index i.
			a.K8S[len(a.K8S)-1] = nil      // Erase last element (write zero value).
			a.K8S = a.K8S[:len(a.K8S)-1]   // Truncate slice.
			log.Printf("k8S %v has been removed from account %v", name, a.Name)
			return nil
		}
	}

	return fmt.Errorf("K8S %v not found", name)
}

func (a *Account) IterK8S() <-chan *k8s.K8S {
	c := make(chan *k8s.K8S)

	f := func() {
		a.lockerK8SSlice.Lock()
		defer a.lockerK8SSlice.Unlock()

		for _, v := range a.K8S {
			c <- v
		}
		close(c)
	}
	go f()

	return c
}

//Software part
func (a *Account) GetNumbersOfSoftware() int {

	a.lockerSoftwareSlice.Lock()
	defer a.lockerSoftwareSlice.Unlock()

	return len(a.Software)
}

func (a *Account) GetSoftwareNameList() []string {

	a.lockerSoftwareSlice.Lock()
	defer a.lockerSoftwareSlice.Unlock()

	softwareNames := make([]string, 0)
	for _, v := range a.Software {
		softwareNames = append(softwareNames, v.Name)
	}

	return softwareNames
}

func (a *Account) GetSoftwareByName(name string) (*saas.Software, error) {

	a.lockerSoftwareSlice.Lock()
	defer a.lockerSoftwareSlice.Unlock()

	for _, v := range a.Software {
		if v.Name == name {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Software %v not found", name)
}

func (a *Account) AppendSoftware(software *saas.Software) {

	a.lockerSoftwareSlice.Lock()
	defer a.lockerSoftwareSlice.Unlock()

	a.Software = append(a.Software, software)
}

func (a *Account) RemoveSoftwareByName(name string) error {

	a.lockerSoftwareSlice.Lock()
	defer a.lockerSoftwareSlice.Unlock()

	//To remove item from slice, this is a fast version (changes order)
	for i, v := range a.Software {
		if v.Name == name {
			// Remove the element at index i from a.
			a.Software[i] = a.Software[len(a.Software)-1] // Copy last element to index i.
			a.Software[len(a.Software)-1] = nil           // Erase last element (write zero value).
			a.Software = a.Software[:len(a.Software)-1]   // Truncate slice.
			log.Printf("Software %v has been removed from account %v", name, a.Name)
			return nil
		}
	}

	return fmt.Errorf("Software %v not found", name)
}

func (a *Account) IterSoftware() <-chan *saas.Software {
	c := make(chan *saas.Software)

	f := func() {
		a.lockerSoftwareSlice.Lock()
		defer a.lockerSoftwareSlice.Unlock()

		for _, v := range a.Software {
			c <- v
		}
		close(c)
	}
	go f()

	return c
}

//Send notification
func (a *Account) SendNotification(msg string) {
	if a.Contract != "" {
		notification.SendNotification(notification.Message{Target: a.Contract, Text: msg})
	}
}

//Set one-time password
//flag -> true, set a random password
//flag -> false, clear this password, set to ""
func (a *Account) SetOneTimePass(flag bool) {
	if flag {
		a.OneTimePass = auth.OneTimePassGen(a.Contract)
		log.Printf("One-time password %v generated for account %v", a.OneTimePass, a.Name)
	} else {
		a.OneTimePass = ""
	}
}

//Get one-time password
func (a *Account) GetOneTimePass() string {
	return a.OneTimePass
}
