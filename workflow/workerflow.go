package workflow

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/vm"
)

// VM live status retry times and interval(unit seconds) setting
const VmStatusRetry, vmStatusInterval = 10, 6

// Create VMs
// Args:
//   account: account
//   vmRequest: vm request body
func CreateVMs(myaccount *account.Account, vmRequest vm.VmRequest) {
	myaccount.StatusVm = "running"
	defer func() {
		myaccount.StatusVm = "idle"
	}()

	//Fetch all existing VM Name, get the last index as the start index of new VM
	vmNames := myaccount.GetVmNameList()
	log.Println(vmNames)
	vmIndex := func(v []string) []string {
		indexes := make([]string, 0)
		for _, v := range v {
			t := strings.Split(v, "-")
			indexes = append(indexes, t[len(t)-1])
		}
		return indexes
	}(vmNames)
	sort.Strings(vmIndex)
	lastIndex := 0
	if len(vmIndex) > 0 {
		lastIndex, _ = strconv.Atoi(vmIndex[len(vmIndex)-1])
		log.Printf("The last index of VM: %v", lastIndex)
	}

	// Create VM in parallel
	log.Printf("VM creation start, numbers: %v", vmRequest.Number)
	var wg sync.WaitGroup
	for i := lastIndex + 1; i <= lastIndex+vmRequest.Number; i++ {
		wg.Add(1)
		go func(i int) {
			//task1: Create VM
			defer wg.Done()
			newVm := vm.NewVirtualMachine(
				vmRequest.Account+"-"+strconv.Itoa(i),
				vmRequest.Flavor,
				vmRequest.Type,
				1, 2048, 20, // if flavor not give, use cpu: 1, mem: 2048M, disk: 20G as default
				node.ComputeNode{IpAddress: "127.0.0.1", UserName: "root", Passwd: "Cisco123!"},
				time.Hour*24*time.Duration(vmRequest.Duration),
			)

			if newVm != nil {
				myaccount.VM = append(myaccount.VM, newVm)
				db.NotifyToDB("account", newVm.Name)
			} else {
				log.Println("Create VM failed, return")
				return
			}

			//task2: Get VM Info
			retry := 1
			for retry <= VmStatusRetry {
				if err := newVm.GetVirtualMachineLiveStatus(); err == nil {
					if newVm.Status != "" && newVm.IpAddress != "" {
						log.Printf("Get new VM -> %v info: status -> %v, address -> %v", newVm.Name, newVm.Status, newVm.IpAddress)
						db.NotifyToDB("account", newVm.Name)
						break
					}
				}
				log.Println("VM get live status failed or empty, will try again")
				time.Sleep(time.Second * vmStatusInterval)
				retry++
			}
			if retry > VmStatusRetry {
				log.Println("VM get status timeout, return")
				return
			}

			//task3: Setup Proxy
			//task4: Send Notification
		}(i)
	}
	wg.Wait()
	log.Println("VM creation done")
}

// Take specify action on VM(start/delete/shutdown/reboot)
// Args:
//   account pointer, vm pointer
//   action -> start/shutdown/reboot/delete
// Return:
//   error -> action error messages
//   nil -> action success
func ActionVM(myAccount *account.Account, myVM *vm.VirtualMachine, action string) error {

	var action_err error
	switch action {
	case "start":
		action_err = myVM.StartUpVirtualMachine()
	case "shutdown":
		action_err = myVM.ShutDownVirtualMachine()
	case "reboot":
		action_err = myVM.RebootVirtualMachine()
	case "delete":
		action_err = myVM.DeleteVirtualMachine()
		//TODO remove from proxy
		if action_err == nil {
			if err := myAccount.RemoveVmByName(myVM.Name); err != nil {
				log.Println(err)
			}
		}
		db.NotifyToDB("account", myVM.Name)
	}

	// Post action, sync up vm status
	switch action {
	case "start", "shutdown", "reboot":
		go func() {
			time.Sleep(time.Second * 3)
			if err := myVM.GetVirtualMachineLiveStatus(); err != nil {
				log.Printf("sync up vm -> %v status after action -> %v, failed -> %v", myVM.Name, action, err)
			}
			db.NotifyToDB("account", myVM.Name)
		}()
	}

	return action_err
}
