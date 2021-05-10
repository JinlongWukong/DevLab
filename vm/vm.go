package vm

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/utils"
)

func NewVirtualMachine(name, flavor, vm_type string, cpu, mem, disk int, nodeName string, Duration time.Duration) *VirtualMachine {

	log.Printf("Creating vm %v on Host %v", name, nodeName)

	//There is a mapping bt flavor and cpu/memory
	detail, exists := flavorDetails[flavor]
	if exists == true {
		cpu = detail["cpu"]
		mem = detail["memory"]
		disk = detail["disk"]
	}

	vnc := VncInfo{
		Port: "unknow",
		Pass: utils.RandomString(8),
	}

	mynode := node.GetNodeByName(nodeName)
	if mynode == nil {
		log.Printf("Error: Node %v not found", nodeName)
		return nil
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"vmName":   name,
		"vmAction": "create",
		"vmMemory": mem,
		"vmVcpus":  cpu,
		"vmDisk":   disk,
		"vmType":   vm_type,
		"vncPass":  vnc.Pass,
		"hostIp":   mynode.IpAddress,
		"hostPass": mynode.Passwd,
		"hostUser": mynode.UserName,
	})

	log.Println("Remote http call to create vm")
	err, _ := utils.HttpSendJsonData("http://10.124.44.167:9134/vm", "POST", payload)
	if err != nil {
		log.Println(err)
		return nil
	}

	ipadd, status := "unknow", "unknow"
	return &VirtualMachine{name, cpu, mem, disk, ipadd, status, vnc, vm_type, nodeName, Duration}
}

// Generic action(start/delete/shutdown/reboot)
// Args:
//    vm     -> VirtualMachine
//    action -> action
// Return:
//    nil    -> success
//    error  -> failed
func (myvm VirtualMachine) genericActionVirtualMachine(action string) error {

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
	err, _ := utils.HttpSendJsonData("http://10.124.44.167:9134/vm", "POST", payload)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (myvm VirtualMachine) DeleteVirtualMachine() error {

	log.Printf("Deleting vm %v on Host %v", myvm.Name, myvm.Node)

	return myvm.genericActionVirtualMachine("delete")
}

func (myvm VirtualMachine) StartUpVirtualMachine() error {

	log.Printf("Starting vm %v on Host %v", myvm.Name, myvm.Node)

	return myvm.genericActionVirtualMachine("start")
}

func (myvm VirtualMachine) ShutDownVirtualMachine() error {

	log.Printf("Shuting down vm %v on Host %v", myvm.Name, myvm.Node)

	return myvm.genericActionVirtualMachine("shutdown")
}

func (myvm VirtualMachine) RebootVirtualMachine() error {

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
	err, reponse_data := utils.HttpGetJsonData("http://10.124.44.167:9134/vm", query)
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
