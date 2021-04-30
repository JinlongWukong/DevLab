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
	err := utils.HttpSendJsonData("http://xxxxxx:9134/vm", "POST", payload)
	if err != nil {
		log.Println(err)
		return nil
	}

	ipadd, status := "", "unknow"
	return &VirtualMachine{name, cpu, mem, disk, ipadd, status, vm_type, host, Duration}
}

func DeleteVirtualMachine(name string, host node.ComputeNode) bool {

	log.Printf("Deleting vm %v on Host %v", name, host.Name)

	return true
}

func GetVirtualMachineLiveStatus(name string, host node.ComputeNode) VmLiveStatus {

	log.Printf("Fetching vm  %v status on Host %v", name, host.Name)

	//vmName=test\&hostIp=127.0.0.1\&hostPass=xxxxx\&hostUser=root
	query := map[string]string{
		"vmName":   name,
		"hostIp":   host.IpAddress,
		"hostUser": host.UserName,
		"hostPass": host.Passwd,
	}

	var vmStatus VmLiveStatus
	err, reponse_data := utils.HttpGetJsonData("http://xxxxxx:9134/vm", query)
	if err == nil {
		json.Unmarshal(reponse_data, &vmStatus)
		return vmStatus
	}

	return vmStatus
}
