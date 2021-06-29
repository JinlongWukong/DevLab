package vm

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/JinlongWukong/CloudLab/deployer"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/utils"
)

//check parameters, return struct pointer if ok, otherwise return nil
func NewVirtualMachine(name, flavor, vmType, hostname, rootPass string, cpu, mem, disk int32, Duration time.Duration) *VirtualMachine {

	//There is a mapping bt flavor and cpu/memory
	if flavor != "" {
		detail, err := GetFlavordetail(flavor)
		if err == nil {
			cpu = detail["cpu"]
			mem = detail["memory"]
			disk = detail["disk"]
		} else {
			log.Println(err)
		}
	}

	if cpu == 0 || mem == 0 || disk == 0 {
		log.Println("Error: one of cpu, mem, disk is zero give")
		return nil
	}

	if hostname == "" {
		hostname = name
	}

	vnc := VncInfo{
		Port: "unknown",
		Pass: utils.RandomString(8),
	}

	if rootPass == "" {
		rootPass = utils.RandomString(8)
	}

	return &VirtualMachine{
		Name:     name,
		Hostname: hostname,
		CPU:      cpu,
		Memory:   mem,
		Disk:     disk,
		Status:   VmStatusInit,
		Vnc:      vnc,
		Type:     vmType,
		Lifetime: Duration,
		PortMap:  map[int]string{},
		RootPass: rootPass,
	}
}

//Create VM by calling remote deployer
func (myvm *VirtualMachine) CreateVirtualMachine() error {

	log.Printf("Creating vm %v on Host %v", myvm.Name, myvm.Node)
	myvm.Status = VmStatusCreating

	node := node.GetNodeByName(myvm.Node)

	payload, _ := json.Marshal(map[string]interface{}{
		"vmName":     myvm.Name,
		"vmHostname": myvm.Hostname,
		"vmAction":   "create",
		"vmMemory":   myvm.Memory,
		"vmVcpus":    myvm.CPU,
		"vmDisk":     myvm.Disk,
		"vmType":     myvm.Type,
		"vncPass":    myvm.Vnc.Pass,
		"rootPass":   myvm.RootPass,
		"hostIp":     node.IpAddress,
		"hostPass":   node.Passwd,
		"hostUser":   node.UserName,
	})

	log.Println("Remote http call to create vm")
	url := deployer.GetDeployerBaseUrl() + "/vm"
	err, _ := utils.HttpSendJsonData(url, "POST", payload)
	if err != nil {
		log.Println(err)
		myvm.Status = fmt.Sprint(err)
		return err
	} else {
		myvm.Status = VmStatusRunning
		return nil
	}
}

// Generic action(start/delete/shutdown/reboot)
// Args:
//    vm     -> VirtualMachine
//    action -> action
// Return:
//    nil    -> success
//    error  -> failed
func (myvm *VirtualMachine) genericActionVirtualMachine(action string) error {

	mynode := node.GetNodeByName(myvm.Node)
	if mynode == nil {
		err := fmt.Errorf("Error: Node %v not found", myvm.Node)
		log.Println(err)
		return err
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"vmName":   myvm.Name,
		"vmAction": action,
		"hostIp":   mynode.IpAddress,
		"hostPass": mynode.Passwd,
		"hostUser": mynode.UserName,
	})

	log.Printf("Remote http call to %v vm", action)
	url := deployer.GetDeployerBaseUrl() + "/vm"
	err, _ := utils.HttpSendJsonData(url, "POST", payload)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (myvm *VirtualMachine) DeleteVirtualMachine() error {

	log.Printf("Deleting vm %v on Host %v", myvm.Name, myvm.Node)

	return myvm.genericActionVirtualMachine("delete")
}

func (myvm *VirtualMachine) StartUpVirtualMachine() error {

	log.Printf("Starting vm %v on Host %v", myvm.Name, myvm.Node)

	return myvm.genericActionVirtualMachine("start")
}

func (myvm *VirtualMachine) ShutDownVirtualMachine() error {

	log.Printf("Shuting down vm %v on Host %v", myvm.Name, myvm.Node)

	return myvm.genericActionVirtualMachine("shutdown")
}

func (myvm *VirtualMachine) RebootVirtualMachine() error {

	log.Printf("Rebooting vm %v on Host %v", myvm.Name, myvm.Node)

	return myvm.genericActionVirtualMachine("reboot")
}

// Sync up VM status
func (myvm *VirtualMachine) GetVirtualMachineLiveStatus() error {

	log.Printf("Fetching vm  %v status on Host %v", myvm.Name, myvm.Node)

	mynode := node.GetNodeByName(myvm.Node)
	if mynode == nil {
		err := fmt.Errorf("Error: Node %v not found", myvm.Node)
		log.Println(err)
		return err
	}

	//vmName=test-1\&hostIp=127.0.0.1\&hostPass=xxxxx\&hostUser=root
	query := map[string]string{
		"vmName":   myvm.Name,
		"hostIp":   mynode.IpAddress,
		"hostUser": mynode.UserName,
		"hostPass": mynode.Passwd,
	}

	var vmStatus VmLiveStatus
	url := deployer.GetDeployerBaseUrl() + "/vm"
	err, reponse_data := utils.HttpGetJsonData(url, query)
	if err != nil {
		log.Println(err)
		return err
	}
	json.Unmarshal(reponse_data, &vmStatus)

	myvm.Status = vmStatus.Status
	myvm.IpAddress = vmStatus.Address
	myvm.Vnc.Port = vmStatus.VncPort
	log.Printf("Fetched vm %v status -> %v, address -> %v, vnc port -> %v", myvm.Name, myvm.Status, myvm.IpAddress, myvm.Vnc.Port)

	return nil
}

func (myvm *VirtualMachine) ActionDnatRule(port []int, action string) error {

	mynode := node.GetNodeByName(myvm.Node)
	if mynode == nil {
		err := fmt.Errorf("Error: Node %v not found", myvm.Node)
		log.Println(err)
		return err
	}

	var rules []map[string]string
	for _, p := range port {
		rules = append(rules, map[string]string{
			"dport":       strings.Split(myvm.PortMap[p], ":")[0],
			"destination": strings.Split(myvm.IpAddress, "/")[0] + ":" + strconv.Itoa(p),
			"state":       action,
			"protocol":    strings.Split(myvm.PortMap[p], ":")[1],
		})
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"rules": rules,
		"Ip":    mynode.IpAddress,
		"Pass":  mynode.Passwd,
		"User":  mynode.UserName,
	})

	log.Printf("Remote http call to %v dnat rule", action)
	url := deployer.GetDeployerBaseUrl() + "/host/dnat"
	err, _ := utils.HttpSendJsonData(url, "POST", payload)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}

func (myvm *VirtualMachine) ChangeLifeTime(delta time.Duration) time.Duration {
	myvm.lifeMutex.Lock()
	defer myvm.lifeMutex.Unlock()

	myvm.Lifetime += delta
	return myvm.Lifetime
}

func (myvm *VirtualMachine) GetLifeTime() time.Duration {
	myvm.lifeMutex.RLock()
	defer myvm.lifeMutex.RUnlock()

	return myvm.Lifetime
}

func GetFlavordetail(flavor string) (map[string]int32, error) {

	//There is a mapping bt flavor and cpu/memory
	detail, exists := flavorDetails[flavor]
	if exists == true {
		return detail, nil
	}

	return nil, fmt.Errorf("flavor not found")
}
