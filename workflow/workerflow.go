package workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/scheduler"
	"github.com/JinlongWukong/CloudLab/utils"
	"github.com/JinlongWukong/CloudLab/vm"
)

var scheduleLock sync.Mutex

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
	var lastIndex int
	if len(vmIndex) > 0 {
		lastIndex, _ = strconv.Atoi(vmIndex[len(vmIndex)-1])
		log.Printf("The last index of VM: %v", lastIndex)
	}

	//schedule for a node
	var reqCpu, reqMem, reqDisk int32
	detail, err := vm.GetFlavordetail(vmRequest.Flavor)
	if err == nil {
		reqCpu = detail["cpu"] * vmRequest.Number
		reqMem = detail["memory"] * vmRequest.Number
		reqDisk = detail["disk"] * 1024 * vmRequest.Number
	} else {
		log.Println(err)
		return
	}
	scheduleLock.Lock()
	selectNode := scheduler.Schedule(reqCpu, reqMem, reqDisk)
	log.Printf("node selected -> %v", selectNode.Name)
	selectNode.ChangeCpuUsed(reqCpu)
	selectNode.ChangeMemUsed(reqMem)
	selectNode.ChangeDiskUsed(reqDisk)
	db.NotifyToDB("node", selectNode.Name, "update")
	scheduleLock.Unlock()

	// Create VM in parallel
	log.Printf("VM creation start, numbers: %v", vmRequest.Number)
	var wg sync.WaitGroup
	for i := lastIndex + 1; i <= lastIndex+int(vmRequest.Number); i++ {
		wg.Add(1)
		go func(i int) {
			//task1: Create VM
			defer wg.Done()
			newVm := vm.NewVirtualMachine(
				vmRequest.Account+"-"+strconv.Itoa(i),
				vmRequest.Flavor,
				vmRequest.Type,
				1, 2048, 20, // if flavor not give, use cpu: 1, mem: 2048M, disk: 20G as default
				selectNode,
				time.Hour*24*time.Duration(vmRequest.Duration),
			)

			if newVm != nil {
				myaccount.VM = append(myaccount.VM, newVm)
				db.NotifyToDB("account", newVm.Name, "create")
			} else {
				log.Println("Create VM failed, return")
				log.Println("return scheduled resources")
				selectNode.ChangeCpuUsed(-(reqCpu / vmRequest.Number))
				selectNode.ChangeMemUsed(-(reqMem / vmRequest.Number))
				selectNode.ChangeDiskUsed(-(reqDisk / vmRequest.Number))
				db.NotifyToDB("node", selectNode.Name, "update")
				return
			}

			//task2: Get VM Info
			retry := 1
			for retry <= VmStatusRetry {
				if err := newVm.GetVirtualMachineLiveStatus(); err == nil {
					if newVm.Status != "" && newVm.IpAddress != "" {
						log.Printf("Get new VM -> %v info: status -> %v, address -> %v", newVm.Name, newVm.Status, newVm.IpAddress)
						db.NotifyToDB("account", newVm.Name, "update")
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
			} else {
				db.NotifyToDB("account", myVM.Name, "delete")
				log.Println("return scheduled resources")
				selectNode := node.GetNodeByName(myVM.Node)
				selectNode.ChangeCpuUsed(-myVM.CPU)
				selectNode.ChangeMemUsed(-myVM.Memory)
				selectNode.ChangeDiskUsed(-myVM.Disk * 1024)
				db.NotifyToDB("node", selectNode.Name, "update")
			}
		}
	}

	// Post action, sync up vm status
	switch action {
	case "start", "shutdown", "reboot":
		go func() {
			time.Sleep(time.Second * 3)
			if err := myVM.GetVirtualMachineLiveStatus(); err != nil {
				log.Printf("sync up vm -> %v status after action -> %v, failed -> %v", myVM.Name, action, err)
			}
			db.NotifyToDB("account", myVM.Name, "update")
		}()
	}

	return action_err
}

// Add a new node
// this is a async call, will update node status after get reponse from remote deployer
func AddNode(nodeRequest node.NodeRequest) error {

	myNode := node.NewNode(nodeRequest)

	_, exists := node.Node_db[nodeRequest.Name]
	if exists == true {
		return fmt.Errorf("node %v already added", nodeRequest.Name)
	} else {
		node.Node_db[myNode.Name] = myNode
		db.NotifyToDB("node", myNode.Name, "create")
	}

	go func() {
		//Install node
		payload, _ := json.Marshal(map[string]interface{}{
			"Ip":     myNode.IpAddress,
			"Pass":   myNode.Passwd,
			"User":   myNode.UserName,
			"Role":   myNode.Role,
			"Action": "install",
		})

		log.Println("Remote http call to install node")
		var nodeInfo node.NodeInfo
		err, reponse_data := utils.HttpSendJsonData("http://10.124.44.167:9134/host", "POST", payload)

		if err != nil {
			log.Printf("Install node  %v failed with error -> %v", myNode.Name, err)
			myNode.Status = node.NodeStatus(fmt.Sprint(err))
			db.NotifyToDB("node", myNode.Name, "update")
			return
		} else {
			log.Printf("Install node %v successfully", myNode.Name)
			json.Unmarshal(reponse_data, &nodeInfo)
			myNode.CPU = nodeInfo.CPU
			myNode.Memory = nodeInfo.Memory
			myNode.Disk = nodeInfo.Disk
			myNode.OSType = nodeInfo.OSType
			log.Printf("Fetched node %v cpu -> %v, memory -> %v, disk -> %v, os type -> %v", myNode.Name, myNode.CPU, myNode.Memory, myNode.Disk, myNode.OSType)
			myNode.Status = node.NodeStatusInstalled
			db.NotifyToDB("node", myNode.Name, "update")
		}
	}()

	return nil
}

// Take specify action on Node(remove/reboot)
// Args:
//   NodeRequest
// Return:
//   error -> action error messages
//   nil -> action success
func ActionNode(nodeRequest node.NodeRequest) error {

	myNode, exists := node.Node_db[nodeRequest.Name]
	if exists == false {
		return fmt.Errorf("node not existed")
	}

	switch nodeRequest.Action {
	case node.NodeActionRemove:
		//TODO ,check whether vm hosted on node
		delete(node.Node_db, nodeRequest.Name)
		db.NotifyToDB("node", nodeRequest.Name, "delete")
		log.Printf("node %v removed", nodeRequest.Name)
	case node.NodeActionReboot:
		if err := myNode.RebootNode(); err != nil {
			log.Printf("Reboot node %v failed with error -> %v", nodeRequest.Name, err)
			return err
		}
	case node.NodeActionEnable:
		myNode.SetState(node.NodeStateEnable)
		log.Printf("Set node %v state to %v", nodeRequest.Name, node.NodeStateEnable)
	case node.NodeActionDisable:
		myNode.SetState(node.NodeStateDisable)
		log.Printf("Set node %v state to %v", nodeRequest.Name, node.NodeStateDisable)
	default:
		return fmt.Errorf("action %v not supported", nodeRequest.Action)
	}

	return nil
}
