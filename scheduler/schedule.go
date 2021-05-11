package scheduler

import (
	"log"
	"math/rand"

	"github.com/JinlongWukong/CloudLab/node"
)

var allocationRatio = 2
var scheduleAlgorithm = "random"

//apply for a node
func Schedule(reqCpu, reqMem, reqDisk int32) *node.Node {

	//filter all nodes
	winerNodes := make([]*node.Node, 0)
	for _, n := range node.Node_db {
		if n.State != node.NodeStateEnable ||
			n.Status != node.NodeStatusInstalled ||
			(n.CPU*int32(allocationRatio)-n.GetCpuUsed()) < reqCpu ||
			(n.Memory*int32(allocationRatio)-n.GetMemUsed()) < reqMem ||
			(n.Disk*int32(allocationRatio)-n.GetDiskUsed()) < reqDisk {
			continue
		} else {
			winerNodes = append(winerNodes, n)
		}
	}

	if len(winerNodes) == 0 {
		log.Println("Not enough nodes left")
		return nil
	}

	//Select one node based on scheduleAlgorithm
	if scheduleAlgorithm == "random" {
		return winerNodes[rand.Intn(len(winerNodes))]
	} else {
		return winerNodes[rand.Intn(len(winerNodes))]
	}

}
