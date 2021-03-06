package workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JinlongWukong/DevLab/account"
	"github.com/JinlongWukong/DevLab/config"
	"github.com/JinlongWukong/DevLab/db"
	"github.com/JinlongWukong/DevLab/deployer"
	"github.com/JinlongWukong/DevLab/k8s"
	"github.com/JinlongWukong/DevLab/node"
	"github.com/JinlongWukong/DevLab/saas"
	"github.com/JinlongWukong/DevLab/scheduler"
	"github.com/JinlongWukong/DevLab/utils"
	"github.com/JinlongWukong/DevLab/vm"
)

var scheduleLock sync.Mutex
var newNodeLock sync.Mutex

// VM live status retry times and interval(unit seconds) setting, 2mins
var vmStatusRetry, vmStatusInterval = 50, 6

// NoVNC related setting
var noVncHost, noVncProtocol = "", ""
var noVncPort int
var noVncUse bool //flag of use noVnc or not

// initialize configuration
func init() {
	if config.Workflow.VmStatusRetry > 0 {
		vmStatusRetry = config.Workflow.VmStatusRetry
	}
	if config.Workflow.VmStatusInterval > 0 {
		vmStatusInterval = config.Workflow.VmStatusInterval
	}
	if config.Workflow.NoVncProtocol != "" &&
		config.Workflow.NoVncHost != "" &&
		config.Workflow.NoVncPort > 0 {
		noVncProtocol = config.Workflow.NoVncProtocol
		noVncHost = config.Workflow.NoVncHost
		noVncPort = config.Workflow.NoVncPort
		noVncUse = true
	}
}

// Create VMs
func CreateVMs(myAccount *account.Account, vmRequest vm.VmRequest) ([]*vm.VirtualMachine, error) {

	myAccount.Lock()
	defer myAccount.Unlock()
	defer db.NotifyToSave()

	//Get the last index as the index of new virtual machine
	lastIndex := utils.GetLastIndex(myAccount.GetVmNameList())

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
			myAccount.Name+"-"+strconv.Itoa(i),
			vmRequest.Flavor,
			vmRequest.Type,
			hostname,
			vmRequest.RootPass,
			vmRequest.CPU,
			vmRequest.Memory,
			vmRequest.Disk,
			time.Hour*24*time.Duration(vmRequest.Duration),
			vmRequest.Addons,
		)
		if newVm != nil {
			myAccount.AppendVM(newVm)
			newVmGroup = append(newVmGroup, newVm)
		} else {
			log.Println("Error: Input paramters not valid")
			return nil, fmt.Errorf("Input paramters not valid")
		}
	}

	go func() {
		changeTaskCount(1)
		defer changeTaskCount(-1)

		//call scheduler to select a node
		reqCpu := newVmGroup[0].CPU * vmRequest.Number
		reqMem := newVmGroup[0].Memory * vmRequest.Number
		reqDisk := newVmGroup[0].Disk * 1024 * vmRequest.Number

		scheduleLock.Lock()
		selectNode := scheduler.Schedule(node.NodeRoleCompute, reqCpu, reqMem, reqDisk)
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
			newVm.NodeAddress = selectNode.IpAddress
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
						// Generate novnc url
						if noVncUse && myVm.NoVnc == "" && myVm.Vnc.Port != "" {
							if vncPort, err := GetVncPort(myVm.Vnc.Port); err == nil {
								myVm.NoVnc = noVncProtocol + "://" + noVncHost + ":" + strconv.Itoa(noVncPort) +
									"/vnc.html?password=" + myVm.Vnc.Pass +
									"&path=vnc/" + selectNode.IpAddress + "/" + vncPort +
									"&autoconnect=true&resize=scale&reconnect=true&show_dot=true"
							} else {
								log.Printf("parse vnc port error: %v %v", myVm.Vnc.Port, err.Error())
							}
						}
						// Tell if ipaddress fetched or not
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

				myAccount.SendNotification(fmt.Sprintf("Your VM %v is running \n\n"+
					"root passwd -> %v\n"+
					"vnc passwd -> %v\n"+
					"vnc endpoint -> %v%v", myVm.Name, myVm.RootPass, myVm.Vnc.Pass, selectNode.IpAddress, myVm.Vnc.Port))

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
				myAccount.SendNotification(fmt.Sprintf("Your VM %v is ready to login using ssh %v -p %v ", myVm.Name, selectNode.IpAddress, sshPort))

				//task4: Install addons
				if len(myVm.Addons) != 0 {
					//addons install retry times
					const ADDONSRETRY = 3
					go func(myVm *vm.VirtualMachine) {
						retry := 1
						for retry <= ADDONSRETRY {
							err = myVm.InstallAddons()
							if err != nil {
								err_msg := fmt.Sprintf("Addons %v on vm %v installation failed with error -> %v", myVm.Addons, myVm.Name, err.Error())
								log.Println(err_msg)
								log.Printf("addons install retry %v", retry)
								if retry >= ADDONSRETRY {
									myAccount.SendNotification(err_msg)
								}
							} else {
								success_msg := fmt.Sprintf("Congratulations! addons %v on vm %v installation succeed", myVm.Addons, myVm.Name)
								log.Println(success_msg)
								myAccount.SendNotification(success_msg)
								break
							}
							retry++
						}
						return
					}(myVm)
				}
			}(newVm)

		}
		wg.Wait()
		log.Println("VM creation done")
	}()

	return newVmGroup, nil
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
		myVM.Status = vm.VmStatusDeleting
		selectNode := node.GetNodeByName(myVM.Node)
		if selectNode != nil {
			//Delete VM from node
			action_err = myVM.DeleteVirtualMachine()
			if action_err != nil {
				log.Printf("Delete vm %v failed", myVM.Name)
				return action_err
			}

			//Clear dnat rules
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
			log.Println("Recycle node resources")
			selectNode.ChangeCpuUsed(-myVM.CPU)
			selectNode.ChangeMemUsed(-myVM.Memory)
			selectNode.ChangeDiskUsed(-myVM.Disk * 1024)
			myVM.Node = ""
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
			time.Sleep(time.Second * 10)
			if err := myVM.GetVirtualMachineLiveStatus(); err != nil {
				log.Printf("sync up vm -> %v status after action -> %v, failed -> %v", myVM.Name, action, err)
			}
		}()
	}

	return action_err
}

func ExtendVMLifetime(myVM *vm.VirtualMachine, period time.Duration) error {
	changeTaskCount(1)
	defer changeTaskCount(-1)
	defer db.NotifyToSave()

	myVM.ChangeLifeTime(period)

	return nil
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

	if _, existed := myVM.PortMap[port]; existed == true {
		msg := fmt.Sprintf("Port %v already exposed", port)
		log.Println(msg)
		return fmt.Errorf(msg)
	}

	myNode := node.GetNodeByName(myVM.Node)
	if myNode == nil {
		return fmt.Errorf("Error: vm %v hosted node %v not found", myVM.Name, myVM.Node)
	}
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
		myNode.SetStatus(node.NodeStatusInstalling)
		db.NotifyToSave()
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
			log.Printf("Install node  %v failed with error -> %v %v", myNode.Name, err, string(reponse_data))
			myNode.SetStatus(node.NodeStatusInstallFailed)
			return
		} else {
			log.Printf("Install node %v successfully", myNode.Name)
			json.Unmarshal(reponse_data, &nodeInfo)
			myNode.CPU = nodeInfo.CPU
			myNode.Memory = nodeInfo.Memory
			myNode.Disk = nodeInfo.Disk
			myNode.OSType = nodeInfo.OSType
			log.Printf("Created node %v info: cpu -> %v, memory -> %v, disk -> %v, os type -> %v", myNode.Name, myNode.CPU, myNode.Memory, myNode.Disk, myNode.OSType)
			myNode.SetStatus(node.NodeStatusInstalled)
		}
	}()

	return nil
}

// Take specify action on Node(remove/reboot)
func ActionNode(name string, action node.NodeAction) error {
	changeTaskCount(1)
	defer changeTaskCount(-1)
	defer db.NotifyToSave()

	myNode, exists := node.NodeDB.Get(name)
	if exists == false {
		return fmt.Errorf("node not existed")
	}

	switch action {
	case node.NodeActionRemove:
		if myNode.GetCpuUsed() > 0 {
			return fmt.Errorf("Still have vm hosted on node %v, can't be removed", name)
		}
		node.NodeDB.Del(name)
		log.Printf("node %v removed", name)
	case node.NodeActionReboot:
		if err := myNode.RebootNode(); err != nil {
			log.Printf("Reboot node %v failed with error -> %v", name, err)
			return err
		}
	case node.NodeActionEnable:
		myNode.SetState(node.NodeStateEnable)
		log.Printf("Set node %v state to %v", name, node.NodeStateEnable)
	case node.NodeActionDisable:
		myNode.SetState(node.NodeStateDisable)
		log.Printf("Set node %v state to %v", name, node.NodeStateDisable)
	default:
		return fmt.Errorf("action %v not supported", action)
	}

	return nil
}

//Create k8s cluster
func CreateK8S(myAccount *account.Account, k8sRequest k8s.K8sRequest) error {

	myAccount.Lock()
	defer myAccount.Unlock()
	defer db.NotifyToSave()

	//Get the last index as the index of new k8s
	lastIndex := utils.GetLastIndex(myAccount.GetK8sNameList())

	newK8s := k8s.NewK8s(myAccount.Name+"-k8s-"+strconv.Itoa(lastIndex+1), k8sRequest)
	if newK8s != nil {
		myAccount.AppendK8S(newK8s)
	} else {
		return fmt.Errorf("Input paramters not valid")
	}

	log.Printf("K8S cluster %v creating ...", newK8s.Name)
	go func() {
		changeTaskCount(1)
		defer changeTaskCount(-1)
		defer db.NotifyToSave()

		//task1: VM instantiation
		newK8s.SetStatus(k8s.K8sStatusBootingVm)

		var flavor = "middle"
		if newK8s.NumOfWorker > 5 {
			flavor = "large"
		}
		vmRequest := vm.VmRequest{
			Hostname: newK8s.Name,
			Type:     "centos7",
			Flavor:   flavor,
			Number:   1,
			Duration: int(newK8s.Lifetime),
		}
		vmGroup, err := CreateVMs(myAccount, vmRequest)
		if err != nil || len(vmGroup) != 1 {
			log.Println("k8s vm creation failed")
			return
		}

		hostVm := vmGroup[0]
		//binding vm and k8s
		newK8s.HostVm = hostVm.Name
		//make sure vm is active before k8s installation
		retry := 1
		for retry <= vmStatusRetry {
			_, _, err = net.ParseCIDR(hostVm.IpAddress)
			if err == nil {
				break
			}
			log.Println("k8s vm ipaddress not assigned, will try again")
			time.Sleep(time.Second * time.Duration(vmStatusInterval))
			retry++
		}
		if retry > vmStatusRetry {
			newK8s.SetStatus(k8s.K8sStatusBootVmFailed)
			log.Println("k8s vm get ipaddress timeout, exited")
			return
		}

		//task2: K8S installation
		newK8s.SetStatus(k8s.K8sStatusInstalling)

		hostNode := node.GetNodeByName(hostVm.Node)
		if hostNode == nil {
			log.Printf("Error: vm %v hosted node %v not found", hostVm.Name, hostVm.Node)
			return
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"Ip":         hostNode.IpAddress,
			"Port":       strings.Split(hostVm.PortMap[22], ":")[0],
			"Pass":       hostVm.RootPass,
			"User":       "root",
			"Controller": newK8s.NumOfContronller,
			"Worker":     newK8s.NumOfWorker,
		})

		//install k8s take long time,so better to notify db to save before activity
		db.NotifyToSave()

		log.Println("Remote http call to install k8s cluster")
		url := deployer.GetDeployerBaseUrl() + "/k8s"
		err, reponse_data := utils.HttpSendJsonData(url, "POST", payload)
		if err != nil {
			newK8s.SetStatus(k8s.K8sStatusInstallFailed)
			err_msg := fmt.Sprintf("k8s cluster %v installation failed with error -> %v %v", newK8s.Name, err, string(reponse_data))
			log.Printf(err_msg)
			myAccount.SendNotification(err_msg)
			return
		} else {
			newK8s.SetStatus(k8s.K8sStatusRunning)
			log.Printf("k8s cluster %v installation successfully", newK8s.Name)
		}

		//task3, send notification
		myAccount.SendNotification(fmt.Sprintf("Your k8s cluster %v is ready to use, Please login vm %v to access cluster", newK8s.Name, hostVm.Name))
	}()

	return nil
}

//Delete k8s cluster
func DeleteK8S(myaccount *account.Account, name string) error {

	changeTaskCount(1)
	defer changeTaskCount(-1)
	defer db.NotifyToSave()

	myk8s, err := myaccount.GetK8sByName(name)
	if err != nil {
		return fmt.Errorf("k8s not found")
	}

	myk8s.SetStatus(k8s.K8sStatusDeleting)

	myvm, err := myaccount.GetVmByName(myk8s.HostVm)
	if err != nil {
		log.Printf("hostvm %v not found", myk8s.HostVm)
	} else {
		if err = ActionVM(myaccount, myvm, "delete"); err != nil {
			return fmt.Errorf("k8s %v hostvm %v delete failed", myk8s.Name, myk8s.HostVm)
		}
	}

	if err = myaccount.RemoveK8sByName(name); err != nil {
		log.Printf("Remove k8s failed with error: %v", err)
		return err
	}

	log.Printf("k8s cluster %v removed successfully", myk8s.Name)
	return nil

}

//Create software
func CreateSoftware(myAccount *account.Account, softwareRequest saas.SoftwareRequest) error {
	myAccount.Lock()
	defer myAccount.Unlock()
	defer db.NotifyToSave()

	//Get the last index as the index of new software
	lastIndex := utils.GetLastIndex(myAccount.GetSoftwareNameList())

	newSoftware := saas.NewSoftware(myAccount.Name+"-"+softwareRequest.Kind+"-"+strconv.Itoa(lastIndex+1), softwareRequest)
	if newSoftware != nil {
		myAccount.AppendSoftware(newSoftware)
	} else {
		return fmt.Errorf("Software request may wrong, create new software failed")
	}

	log.Printf("Software %v creating ...", newSoftware.Name)
	go func() {
		changeTaskCount(1)
		defer changeTaskCount(-1)
		newSoftware.Lock()
		defer newSoftware.Unlock()
		defer db.NotifyToSave()

		//task1: Software installation
		if newSoftware.Backend == "container" {

			//call scheduler to select a node
			reqCpu := newSoftware.CPU
			reqMem := newSoftware.Memory
			scheduleLock.Lock()
			selectNode := scheduler.Schedule(node.NodeRoleContainer, int32(reqCpu), int32(reqMem), 0)
			if selectNode == nil {
				err_msg := fmt.Sprintf("Error: No valid node selected, software %v creation exit", newSoftware.Name)
				log.Printf(err_msg)
				myAccount.SendNotification(err_msg)
				scheduleLock.Unlock()
				return
			}
			log.Printf("node selected -> %v for software %v", selectNode.Name, newSoftware.Name)
			selectNode.ChangeCpuUsed(int32(reqCpu))
			selectNode.ChangeMemUsed(int32(reqMem))
			selectNode.ChangeDiskUsed(0)
			scheduleLock.Unlock()

			newSoftware.Node = selectNode.Name
			newSoftware.SetStatus(saas.SoftwareStatusScheduled)
			newSoftware.SetStatus(saas.SoftwareStatusInstalling)
			payload, _ := json.Marshal(map[string]interface{}{
				"Ip":       selectNode.IpAddress,
				"Pass":     selectNode.Passwd,
				"User":     selectNode.UserName,
				"Name":     newSoftware.Name,
				"Software": newSoftware.Kind,
				"Version":  newSoftware.Version,
				"Cpu":      newSoftware.CPU,
				"Memory":   strconv.Itoa(int(newSoftware.Memory)) + "m",
			})

			log.Printf("Remote http call to install software %v", newSoftware.Name)
			url := deployer.GetDeployerBaseUrl() + "/container"
			err, reponse_data := utils.HttpSendJsonData(url, "POST", payload)
			if err != nil {
				newSoftware.SetStatus(saas.SoftwareStatusInstallFailed)
				err_msg := fmt.Sprintf("software %v installation failed with error -> %v %v", newSoftware.Name, err, string(reponse_data))
				log.Printf(err_msg)
				myAccount.SendNotification(err_msg)
				return
			} else {
				log.Printf("software %v installation successfully", newSoftware.Name)
				readContainerStatus(newSoftware, reponse_data)
			}
		} else {
			newSoftware.SetStatus(saas.SoftwareStatusError)
			err_msg := fmt.Sprintf("%v backend not support", newSoftware.Backend)
			log.Printf(err_msg)
			myAccount.SendNotification(err_msg)
			return
		}

		//task2, send notification
		myAccount.SendNotification(fmt.Sprintf("Your software %v is created", newSoftware.Name))
	}()

	return nil
}

func ActionSoftware(myAccount *account.Account, name string, action saas.SoftwareAction) error {
	changeTaskCount(1)
	defer changeTaskCount(-1)
	defer db.NotifyToSave()

	mySoftware, err := myAccount.GetSoftwareByName(name)
	if err != nil {
		log.Printf("Software action failed with error: %v", err)
		return err
	}

	mySoftware.Lock()
	defer mySoftware.Unlock()

	selectNode := node.GetNodeByName(mySoftware.Node)
	if selectNode == nil {
		return fmt.Errorf("Error: software %v hosted node %v not found", mySoftware.Name, mySoftware.Node)
	}

	if mySoftware.Backend == "container" {
		payload, _ := json.Marshal(map[string]interface{}{
			"Ip":       selectNode.IpAddress,
			"Pass":     selectNode.Passwd,
			"User":     selectNode.UserName,
			"Name":     mySoftware.Name,
			"Software": mySoftware.Kind,
			"Action":   action,
		})
		url := deployer.GetDeployerBaseUrl() + "/container/action"

		log.Printf("Remote http call to %v software %v", action, name)
		err, reponse_data := utils.HttpSendJsonData(url, "POST", payload)
		if err != nil {
			mySoftware.SetStatus(saas.SoftwareStatusError)
			log.Printf("Remote http call to %v software %v failed with error: %v %v", action, name, err, string(reponse_data))
			myAccount.SendNotification(fmt.Sprintf("%v your software %v failed with error: %v %v", action, mySoftware.Name, err, string(reponse_data)))
			return err
		}

		switch action {
		case saas.SoftwareActionStart, saas.SoftwareActionRestart, saas.SoftwareActionGet:
			readContainerStatus(mySoftware, reponse_data)
		case saas.SoftwareActionStop:
			mySoftware.Address = ""
			mySoftware.PortMapping = nil
			mySoftware.SetStatus(saas.SoftwareStatusStopped)
		default:
			return fmt.Errorf("Software action %v not supported", action)
		}

	}

	log.Printf("%v software %v successfully", action, mySoftware.Name)
	return nil
}

func DeleteSoftware(myAccount *account.Account, name string) error {
	changeTaskCount(1)
	defer changeTaskCount(-1)
	defer db.NotifyToSave()

	mySoftware, err := myAccount.GetSoftwareByName(name)
	if err != nil {
		log.Printf("Software action failed with error: %v", err)
		return err
	}

	mySoftware.Lock()
	defer mySoftware.Unlock()

	mySoftware.SetStatus(saas.SoftwareStatusDeleting)

	action := saas.SoftwareActionDelete
	if mySoftware.Backend == "container" {
		//Recycle node resouces
		selectNode := node.GetNodeByName(mySoftware.Node)
		if selectNode != nil {
			payload, _ := json.Marshal(map[string]interface{}{
				"Ip":       selectNode.IpAddress,
				"Pass":     selectNode.Passwd,
				"User":     selectNode.UserName,
				"Name":     mySoftware.Name,
				"Software": mySoftware.Kind,
				"Action":   action,
			})
			url := deployer.GetDeployerBaseUrl() + "/container/action"

			log.Printf("Remote http call to %v software %v", action, name)
			err, reponse_data := utils.HttpSendJsonData(url, "POST", payload)
			if err != nil {
				mySoftware.SetStatus(saas.SoftwareStatusError)
				log.Printf("Remote http call to %v software %v failed with error: %v %v", action, name, err, string(reponse_data))
				myAccount.SendNotification(fmt.Sprintf("%v your software %v failed with error: %v %v", action, mySoftware.Name, err, string(reponse_data)))
				return err
			}

			log.Println("Recycle node resources")
			selectNode.ChangeCpuUsed(-int32(mySoftware.CPU))
			selectNode.ChangeMemUsed(-int32(mySoftware.Memory))
			selectNode.ChangeDiskUsed(0)
			mySoftware.Node = ""
		}

		if err = myAccount.RemoveSoftwareByName(mySoftware.Name); err != nil {
			log.Printf("Delete software %v failed with error: %v", mySoftware.Name, err)
			return err
		}
	}

	log.Printf("Delete software %v successfully", mySoftware.Name)
	return nil
}
