package scheduler

import (
	"log"
	"math"
	"math/rand"
	"sort"

	"github.com/JinlongWukong/DevLab/config"
	"github.com/JinlongWukong/DevLab/node"
)

var allocationRatio = 2
var scheduleAlgorithm = "weight"

//initialize configuration
func init() {
	if config.Schedule.AllocationRatio > 0 {
		allocationRatio = config.Schedule.AllocationRatio
	}
	if config.Schedule.ScheduleAlgorithm != "" {
		scheduleAlgorithm = config.Schedule.ScheduleAlgorithm
	}
}

//apply for a node
func Schedule(role node.NodeRole, reqCpu, reqMem, reqDisk int32) *node.Node {
	//filter all nodes
	winerNodes := make([]*node.Node, 0)
	for n := range node.NodeDB.Iter() {
		if role != n.Value.Role {
			continue
		} else if n.Value.State != node.NodeStateEnable ||
			n.Value.Status != node.NodeStatusReady ||
			(n.Value.CPU*int32(allocationRatio)-n.Value.GetCpuUsed()) < reqCpu ||
			(n.Value.Memory*int32(allocationRatio)-n.Value.GetMemUsed()) < reqMem ||
			(n.Value.Disk*int32(allocationRatio)-n.Value.GetDiskUsed()) < reqDisk {
			continue
		} else {
			winerNodes = append(winerNodes, n.Value)
		}
	}
	if len(winerNodes) == 0 {
		log.Println("No available node left")
		return nil
	}
	//Select one node based on scheduleAlgorithm
	if scheduleAlgorithm == "random" {
		return winerNodes[rand.Intn(len(winerNodes))]
	} else if scheduleAlgorithm == "weight" {
		return weightSelector(winerNodes)
	} else {
		return winerNodes[rand.Intn(len(winerNodes))]
	}
}

//Select a node which have biggest weight
//weight=the percent of cpu left*100 + the percent of mem left*100 + the percent of disk left*100
func weightSelector(filtedNodes []*node.Node) *node.Node {
	nMap := map[float64]*node.Node{}
	keys := make([]float64, 0)
	for _, n := range filtedNodes {
		cpu := n.CPU * int32(allocationRatio)
		mem := n.Memory * int32(allocationRatio)
		disk := n.Disk * int32(allocationRatio)
		cpuWeight := float64(cpu-n.GetCpuUsed()) / float64(cpu) * 100
		memWeight := float64(mem-n.GetMemUsed()) / float64(mem) * 100
		diskWeight := float64(disk-n.GetDiskUsed()) / float64(disk) * 100
		weight := cpuWeight + memWeight + diskWeight
		roundWeight := math.Round(weight*100) / 100
		nMap[roundWeight] = n
		keys = append(keys, roundWeight)
	}
	sort.Float64s(keys)
	log.Println(keys)
	return nMap[keys[len(keys)-1]]
}
