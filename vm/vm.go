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
func genericActionVirtualMachine(vm VirtualMachine, action string) error {

	payload, _ := json.Marshal(map[string]interface{}{
		"vmName":   vm.Name,
		"vmAction": action,
		"hostIp":   vm.Host.IpAddress,
		"hostPass": vm.Host.Passwd,
		"hostUser": vm.Host.UserName,
	})

	log.Printf("Remote http call to %v vm", action)
	err := utils.HttpSendJsonData("http://10.124.44.167:9134/vm", "POST", payload)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func DeleteVirtualMachine(vm VirtualMachine) error {

	log.Printf("Deleting vm %v on Host %v", vm.Name, vm.Host.Name)

	return genericActionVirtualMachine(vm, "delete")
}

func StartUpVirtualMachine(vm VirtualMachine) error {

	log.Printf("Starting vm %v on Host %v", vm.Name, vm.Host.Name)

	return genericActionVirtualMachine(vm, "start")
}

func ShutDownVirtualMachine(vm VirtualMachine) error {

	log.Printf("Shuting down vm %v on Host %v", vm.Name, vm.Host.Name)

	return genericActionVirtualMachine(vm, "shutdown")
}

func RebootVirtualMachine(vm VirtualMachine) error {

	log.Printf("Rebooting vm %v on Host %v", vm.Name, vm.Host.Name)

	return genericActionVirtualMachine(vm, "reboot")
}

func GetVirtualMachineLiveStatus(vm VirtualMachine) VmLiveStatus {

	log.Printf("Fetching vm  %v status on Host %v", vm.Name, vm.Host.Name)

	//vmName=test\&hostIp=127.0.0.1\&hostPass=xxxxx\&hostUser=root
	query := map[string]string{
		"vmName":   vm.Name,
		"hostIp":   vm.Host.IpAddress,
		"hostUser": vm.Host.UserName,
		"hostPass": vm.Host.Passwd,
	}

	var vmStatus VmLiveStatus
	err, reponse_data := utils.HttpGetJsonData("http://10.124.44.167:9134/vm", query)
	if err == nil {
		json.Unmarshal(reponse_data, &vmStatus)
		return vmStatus
	}

	return vmStatus
}
