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
	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/deployer"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/scheduler"
	"github.com/JinlongWukong/CloudLab/utils"
	"github.com/JinlongWukong/CloudLab/vm"
)

var scheduleLock sync.Mutex
var newNodeLock sync.Mutex

// VM live status retry times and interval(unit seconds) setting, 2mins
var vmStatusRetry, vmStatusInterval = 20, 6

// initialize configuration
func Initialize() {
	if config.Workflow.VmStatusRetry > 0 {
		vmStatusRetry = config.Workflow.VmStatusRetry
	}
	if config.Workflow.VmStatusInterval > 0 {
		vmStatusInterval = config.Workflow.VmStatusInterval
	}
}

// Create VMs
func CreateVMs(myAccount *account.Account, vmRequest vm.VmRequest) error {

	//No more new create task accept during vm creation
	myAccount.StatusVm = "running"
	defer func() {
		myAccount.StatusVm = "idle"
	}()

	//Fetch all existing VM Name, get the last index as the start index of new VM
	vmNames := myAccount.GetVmNameList()
	log.Println(vmNames)
	vmIndex := func(v []string) []int {
		indexes := make([]int, 0)
		for _, v := range v {
			t := strings.Split(v, "-")
			i, _ := strconv.Atoi(t[len(t)-1])
			indexes = append(indexes, i)
		}
		return indexes
	}(vmNames)
	sort.Ints(vmIndex)
	var lastIndex int
	if len(vmIndex) > 0 {
		lastIndex = vmIndex[len(vmIndex)-1]
		log.Printf("The last index of VM: %v", lastIndex)
	}

	// New VM instance
	log.Printf("VM creation starting... total numbers: %v", vmRequest.Number)
	var newVmGroup []*vm.VirtualMachine
	hostname := vmRequest.Hostname
	for i := lastIndex + 1; i <= lastIndex+int(vmRequest.Number); i++ {

		// define hostname if multi instances, by adding index
		if vmRequest.Number != 1 && vmRequest.Hostname != "" {
			hostname = vmRequest.Hostname + "-" + strconv.Itoa(i-lastIndex)
		}

		newVm := vm.NewVirtualMachine(
			vmRequest.Account+"-"+strconv.Itoa(i),
			vmRequest.Flavor,
			vmRequest.Type,
			hostname,
			vmRequest.RootPass,
			vmRequest.CPU,
			vmRequest.Memory,
			vmRequest.Disk,
			time.Hour*24*time.Duration(vmRequest.Duration),
		)
		if newVm != nil {
			myAccount.AppendVM(newVm)
			newVmGroup = append(newVmGroup, newVm)
		} else {
			return fmt.Errorf("Input paramters not valid")
		}
	}

	db.NotifyToSave()

	go func() {
		changeTaskCount(1)
		defer changeTaskCount(-1)

		//call scheduler to select a node
		reqCpu := newVmGroup[0].CPU * vmRequest.Number
		reqMem := newVmGroup[0].Memory * vmRequest.Number
		reqDisk := newVmGroup[0].Disk * 1024 * vmRequest.Number

		scheduleLock.Lock()
		selectNode := scheduler.Schedule(reqCpu, reqMem, reqDisk)
		if selectNode == nil {
			log.Println("Error: No valid node selected, VM creation exit")
			scheduleLock.Unlock()
			return
		}
		log.Printf("node selected -> %v", selectNode.Name)
		selectNode.ChangeCpuUsed(reqCpu)
		selectNode.ChangeMemUsed(reqMem)
		selectNode.ChangeDiskUsed(reqDisk)
		scheduleLock.Unlock()

		for _, newVm := range newVmGroup {
			newVm.Node = selectNode.Name
			newVm.Status = vm.VmStatusScheduled
		}

		db.NotifyToSave()

		//Create VMs in parallel
		var wg sync.WaitGroup
		for _, newVm := range newVmGroup {
			wg.Add(1)
			go func(myVm *vm.VirtualMachine) {

				defer wg.Done()
				myVm.Lock()
				defer myVm.Unlock()
				defer db.NotifyToSave()

				//task1: VM instantiation
				log.Printf("VM %v instantiation start", myVm.Name)
				err := myVm.CreateVirtualMachine()
				if err == nil {
					log.Printf("VM %v instantiation success", myVm.Name)
				} else {
					log.Printf("VM %v instantiation fail", myVm.Name)
					return
				}

				//task2: Get VM Info
				retry := 1
				for retry <= vmStatusRetry {
					if err := myVm.GetVirtualMachineLiveStatus(); err == nil {
						if myVm.Status != "" && myVm.IpAddress != "" {
							log.Printf("Get VM -> %v info: status -> %v, address -> %v", myVm.Name, myVm.Status, myVm.IpAddress)
							break
						}
					}
					log.Println("VM get live status failed or empty, will try again")
					time.Sleep(time.Second * time.Duration(vmStatusInterval))
					retry++
				}
				if retry > vmStatusRetry {
					log.Println("VM get status timeout, exited")
					return
				}
				myAccount.SendNotification(fmt.Sprintf("Your VM %v is running, root pass-> %v, vnc pass -> %v ", myVm.Name, myVm.RootPass, myVm.Vnc.Pass))

				//task3: Setup DNAT
				sshPort := selectNode.ReservePort(strings.Split(myVm.IpAddress, "/")[0] + ":22")
				if sshPort == 0 {
					log.Printf("No port reserved on node %v", selectNode.Name)
					return
				} else {
					myVm.PortMap[22] = strconv.Itoa(sshPort) + ":tcp"
					log.Printf("port -> %v reserved on node for vm %v", sshPort, myVm.Name)
				}
				err = myVm.ActionDnatRule([]int{22}, "present")
				if err != nil {
					log.Println(err)
					myVm.Status = fmt.Sprint(err)
					return
				}
				log.Printf("DNAT setup success for vm %v, port mapping -> %v:%v", myVm.Name, 22, myVm.PortMap[22])

				//task4: Send Notification
				myAccount.SendNotification(fmt.Sprintf("Your VM %v dnat setup done, ssh -> %v -p %v ", myVm.Name, selectNode.IpAddress, sshPort))
			}(newVm)

		}
		wg.Wait()
		log.Println("VM creation done")
	}()

	return nil
}

// Take specify action on VM(start/delete/shutdown/reboot)
func ActionVM(myAccount *account.Account, myVM *vm.VirtualMachine, action string) error {
	changeTaskCount(1)
	defer changeTaskCount(-1)

	myVM.Lock()
	defer myVM.Unlock()

	defer db.NotifyToSave()

	var action_err error
	switch action {
	case "start":
		//VM status check
		if myVM.Status == vm.VmStatusDeleted || myVM.Status == vm.VmStatusDeleting {
			return fmt.Errorf("VM in deleting or deleted")
		}
		action_err = myVM.StartUpVirtualMachine()
	case "shutdown":
		//VM status check
		if myVM.Status == vm.VmStatusDeleted || myVM.Status == vm.VmStatusDeleting {
			return fmt.Errorf("VM in deleting or deleted")
		}
		action_err = myVM.ShutDownVirtualMachine()
	case "reboot":
		//VM status check
		if myVM.Status == vm.VmStatusDeleted || myVM.Status == vm.VmStatusDeleting {
			return fmt.Errorf("VM in deleting or deleted")
		}
		action_err = myVM.RebootVirtualMachine()
	case "delete":
		//Remove vm from account directly since VM is init status, no further action needed
		if myVM.Status == vm.VmStatusInit {
			if err := myAccount.RemoveVmByName(myVM.Name); err != nil {
				log.Println(err)
				return err
			} else {
				myVM.Status = vm.VmStatusDeleted
				return nil
			}
		}

		myVM.Status = vm.VmStatusDeleting

		//Delete VM from node
		action_err = myVM.DeleteVirtualMachine()
		if action_err != nil {
			log.Printf("Delete vm %v failed", myVM.Name)
			return action_err
		}

		//Clear dnat rules
		selectNode := node.GetNodeByName(myVM.Node)
		keys := make([]int, 0, len(myVM.PortMap))
		for k := range myVM.PortMap {
			keys = append(keys, k)
		}
		if len(keys) > 0 {
			err := myVM.ActionDnatRule(keys, "absent")
			if err != nil {
				log.Printf("Clear dnat for vm %v on host %v failed with error %v", myVM.Name, selectNode.Name, err)
				return err
			} else {
				log.Printf("Clear dnat for vm %v on host %v success", myVM.Name, selectNode.Name)
				for _, info := range myVM.PortMap {
					t := strings.Split(info, ":")
					p, _ := strconv.Atoi(t[0])
					selectNode.ReleasePort(p)
				}
				myVM.PortMap = make(map[int]string)
			}
		}

		//Recycle resouces to node
		if myVM.Status != vm.VmStatusInit {
			log.Println("Recycle node resources")
			selectNode.ChangeCpuUsed(-myVM.CPU)
			selectNode.ChangeMemUsed(-myVM.Memory)
			selectNode.ChangeDiskUsed(-myVM.Disk * 1024)
		}

		//Remove from account
		if err := myAccount.RemoveVmByName(myVM.Name); err != nil {
			log.Println(err)
			return err
		} else {
			myVM.Status = vm.VmStatusDeleted
		}

		return nil
	}

	// Post action, fetch latest vm status
	switch action {
	case "start", "shutdown", "reboot":
		go func() {
			time.Sleep(time.Second * 5)
			if err := myVM.GetVirtualMachineLiveStatus(); err != nil {
				log.Printf("sync up vm -> %v status after action -> %v, failed -> %v", myVM.Name, action, err)
			}
		}()
	}

	return action_err
}

// Set dnat rule to expose vm port with node port
func ExposePort(myAccount *account.Account, myVM *vm.VirtualMachine, port int, protocol string) error {
	changeTaskCount(1)
	defer changeTaskCount(-1)

	myVM.Lock()
	defer myVM.Unlock()

	//VM status check
	if myVM.Status == vm.VmStatusDeleted || myVM.Status == vm.VmStatusDeleting {
		return fmt.Errorf("VM in deleting or deleted")
	}

	//Input param check
	if port <= 0 {
		return fmt.Errorf("invalid port given")
	}
	if protocol != "tcp" && protocol != "udp" {
		return fmt.Errorf("invalid protocol given")
	}

	if _, existed := myVM.PortMap[port]; existed == true {
		msg := fmt.Sprintf("Port %v already exposed", port)
		log.Println(msg)
		return fmt.Errorf(msg)
	}

	myNode := node.GetNodeByName(myVM.Node)
	newPort := myNode.ReservePort(strings.Split(myVM.IpAddress, "/")[0] + ":" + strconv.Itoa(port))

	defer db.NotifyToSave()

	if newPort == 0 {
		msg := fmt.Sprintf("No port reserved on node %v", myNode.Name)
		log.Println(msg)
		return fmt.Errorf(msg)
	} else {
		myVM.PortMap[port] = strconv.Itoa(newPort) + ":" + protocol
		log.Printf("port -> %v reserved on node for vm %v", newPort, myVM.Name)
	}
	err := myVM.ActionDnatRule([]int{port}, "present")
	if err != nil {
		log.Println(err)
		myNode.ReleasePort(port)
		myVM.Status = fmt.Sprint(err)
		return err
	}
	log.Printf("DNAT setup success for vm %v, port mapping -> %v:%v", myVM.Name, port, myVM.PortMap[port])

	return nil

}

// Add a new node
// this is a async call, will update node status after get reponse from remote deployer
func AddNode(nodeRequest node.NodeRequest) error {

	_, exists := node.NodeDB.Get(nodeRequest.Name)
	if exists == true {
		return fmt.Errorf("node %v already added", nodeRequest.Name)
	}

	newNodeLock.Lock()
	myNode := node.NewNode(nodeRequest)
	node.NodeDB.Set(myNode.Name, myNode)
	newNodeLock.Unlock()
	db.NotifyToSave()

	go func() {
		changeTaskCount(1)
		defer changeTaskCount(-1)
		defer db.NotifyToSave()

		//Install node
		payload, _ := json.Marshal(map[string]interface{}{
			"Ip":     myNode.IpAddress,
			"Pass":   myNode.Passwd,
			"User":   myNode.UserName,
			"Role":   myNode.Role,
			"Action": "install",
			"Subnet": myNode.Subnet,
		})

		log.Println("Remote http call to install node")
		var nodeInfo node.NodeInfo
		url := deployer.GetDeployerBaseUrl() + "/host"
		err, reponse_data := utils.HttpSendJsonData(url, "POST", payload)

		if err != nil {
			log.Printf("Install node  %v failed with error -> %v", myNode.Name, err)
			myNode.SetStatus(node.NodeStatusFailed)
			return
		} else {
			log.Printf("Install node %v successfully", myNode.Name)
			json.Unmarshal(reponse_data, &nodeInfo)
			myNode.CPU = nodeInfo.CPU
			myNode.Memory = nodeInfo.Memory
			myNode.Disk = nodeInfo.Disk
			myNode.OSType = nodeInfo.OSType
			log.Printf("Fetched node %v cpu -> %v, memory -> %v, disk -> %v, os type -> %v", myNode.Name, myNode.CPU, myNode.Memory, myNode.Disk, myNode.OSType)
			myNode.SetStatus(node.NodeStatusInstalled)
		}
	}()

	return nil
}

// Take specify action on Node(remove/reboot)
func ActionNode(nodeRequest node.NodeRequest) error {
	changeTaskCount(1)
	defer changeTaskCount(-1)
	defer db.NotifyToSave()

	myNode, exists := node.NodeDB.Get(nodeRequest.Name)
	if exists == false {
		return fmt.Errorf("node not existed")
	}

	switch nodeRequest.Action {
	case node.NodeActionRemove:
		if myNode.GetCpuUsed() > 0 {
			return fmt.Errorf("Still have vm hosted on node %v, can't be removed", myNode.Name)
		}
		node.NodeDB.Del(nodeRequest.Name)
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
