package workflow

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/vm"
)

// VM live status retry times and interval(unit seconds) setting
const VmStatusRetry, vmStatusInterval = 10, 15

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
			} else {
				log.Println("Create VM failed, return")
				return
			}

			//task2: Get VM Info
			retry := 1
			for retry <= VmStatusRetry {
				vmStatus := vm.GetVirtualMachineLiveStatus(*newVm)

				if vmStatus.Name == newVm.Name && vmStatus.Address != "" {
					newVm.IpAddress = vmStatus.Address
					newVm.Status = vmStatus.Status
					log.Printf("Get VM -> %v live status -> %v, address -> %v", vmStatus.Name, vmStatus.Status, vmStatus.Address)
					break
				}

				log.Println("VM Live status empty, will try again")
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
