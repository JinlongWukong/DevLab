package workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/saas"
)

var taskCount int64

func GetTaskCount() int64 {
	return atomic.LoadInt64(&taskCount)
}

func changeTaskCount(delta int64) {
	atomic.AddInt64(&taskCount, delta)
}

func readContainerStatus(mySoftware *saas.Software, reponse_data []byte) error {
	var softwareInfo saas.SoftwareInfo
	if err := json.Unmarshal(reponse_data, &softwareInfo); err == nil {
		mySoftware.Address = softwareInfo.Address
		for k, v := range softwareInfo.AdditionalInfor {
			mySoftware.AdditionalInfor[k] = v
		}
		switch softwareInfo.Status {
		case "running":
			mySoftware.SetStatus(saas.SoftwareStatusRunning)
		case "stopped":
			mySoftware.SetStatus(saas.SoftwareStatusStopped)
		case "deleted":
			mySoftware.SetStatus(saas.SoftwareStatusNotFound)
		case "unknown":
			mySoftware.SetStatus(saas.SoftwareStatusUnknown)
		default:
			mySoftware.SetStatus(saas.SoftwareStatusError)
		}
		mySoftware.PortMapping = map[string]string{}
		softwareNode := node.GetNodeByName(mySoftware.Node)
		if softwareNode == nil {
			return fmt.Errorf("Error: software %v hosted node %v not found", mySoftware.Name, mySoftware.Node)
		}
		for _, v := range softwareInfo.PortMapping {
			format1 := strings.Split(v, "->")
			format2 := strings.Split(v, ":")
			if len(format2) != 2 || len(format1) != 2 {
				log.Printf("invalid format port mapping found %v, skip", v)
				break
			}
			left := format1[0]
			right := format2[1]
			mySoftware.PortMapping[strings.Trim(left, " ")] = softwareNode.IpAddress + ":" + right
		}
	} else {
		log.Printf("Failed to unmarshal software %v status information", mySoftware.Name)
		return err
	}

	return nil
}
