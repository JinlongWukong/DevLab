package vm

import (
	"encoding/json"
	"log"
	"time"

	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/utils"
)

func NewVirtualMachine(name, flavor, vm_type string, cpu, mem, disk int, host node.ComputeNode, Duration time.Duration) *VirtualMachine {

	log.Printf("Creating vm %v on Host %v", name, host.Name)

	//There is a mapping bt flavor and cpu/memory
	detail, exists := flavorDetails[flavor]
	if exists == true {
		cpu = detail["cpu"]
		mem = detail["memory"]
		disk = detail["disk"]
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"vmName":   name,
		"vmAction": "create",
		"vmMemory": mem,
		"vmVcpus":  cpu,
		"vmDisk":   disk,
		"vmType":   vm_type,
		"hostIp":   host.IpAddress,
		"hostPass": host.Passwd,
		"hostUser": host.UserName,
	})

	log.Println("Remote http call to create vm")
	err := utils.HttpSendJsonData("http://10.124.44.167:9134/vm", "POST", payload)
	if err != nil {
		log.Println(err)
		return nil
	}

	ipadd, status := "", "unknow"
	return &VirtualMachine{name, cpu, mem, disk, ipadd, status, vm_type, host, Duration}
}

// Generic action(start/delete/shutdown/reboot)
// Args:
//    vm     -> VirtualMachine
//    action -> action
// Return:
//    nil    -> success
//    error  -> failed
func (myvm VirtualMachine) genericActionVirtualMachine(action string) error {

	payload, _ := json.Marshal(map[string]interface{}{
		"vmName":   myvm.Name,
		"vmAction": action,
		"hostIp":   myvm.Host.IpAddress,
		"hostPass": myvm.Host.Passwd,
		"hostUser": myvm.Host.UserName,
	})

	log.Printf("Remote http call to %v vm", action)
	err := utils.HttpSendJsonData("http://10.124.44.167:9134/vm", "POST", payload)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (myvm VirtualMachine) DeleteVirtualMachine() error {

	log.Printf("Deleting vm %v on Host %v", myvm.Name, myvm.Host.Name)

	return myvm.genericActionVirtualMachine("delete")
}

func (myvm VirtualMachine) StartUpVirtualMachine() error {

	log.Printf("Starting vm %v on Host %v", myvm.Name, myvm.Host.Name)

	return myvm.genericActionVirtualMachine("start")
}

func (myvm VirtualMachine) ShutDownVirtualMachine() error {

	log.Printf("Shuting down vm %v on Host %v", myvm.Name, myvm.Host.Name)

	return myvm.genericActionVirtualMachine("shutdown")
}

func (myvm VirtualMachine) RebootVirtualMachine() error {

	log.Printf("Rebooting vm %v on Host %v", myvm.Name, myvm.Host.Name)

	return myvm.genericActionVirtualMachine("reboot")
}

// Sync up VM status
func (myvm *VirtualMachine) GetVirtualMachineLiveStatus() error {

	log.Printf("Fetching vm  %v status on Host %v", myvm.Name, myvm.Host.Name)

	//vmName=test-1\&hostIp=127.0.0.1\&hostPass=xxxxx\&hostUser=root
	query := map[string]string{
		"vmName":   myvm.Name,
		"hostIp":   myvm.Host.IpAddress,
		"hostUser": myvm.Host.UserName,
		"hostPass": myvm.Host.Passwd,
	}

	var vmStatus VmLiveStatus
	err, reponse_data := utils.HttpGetJsonData("http://10.124.44.167:9134/vm", query)
	if err != nil {
		log.Println(err)
		return err
	}
	json.Unmarshal(reponse_data, &vmStatus)

	myvm.Status = vmStatus.Status
	myvm.IpAddress = vmStatus.Address
	log.Printf("Fetched vm %v status -> %v, address -> %v", myvm.Name, myvm.Status, myvm.IpAddress)

	return nil
}
