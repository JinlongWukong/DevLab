package network

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/manager"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/utils"
)

var checkInterval = "10s"
var networkType = "hostgw"
var nodeSubnetCache = []string{}

type NetworkController struct {
}

var _ manager.Manager = NetworkController{}

//initialize configuration
func initialize() {
	if config.Network.CheckInterval != "" {
		checkInterval = config.Network.CheckInterval
	}
	if config.Network.NetworkType != "" {
		networkType = config.Network.NetworkType
	}
}

func (n NetworkController) Control(ctx context.Context, wg *sync.WaitGroup) {

	log.Println("NetworkController manager started")
	defer func() {
		log.Println("NetworkController manager exited")
		wg.Done()
	}()

	initialize()

	period, err := time.ParseDuration(checkInterval)
	if err != nil {
		log.Println(err)
		return
	}
	t := time.NewTicker(period)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			//Copy all nodes
			allNodes := []*node.Node{}
			for v := range node.NodeDB.Iter() {
				//if node status not installed or failed, skip
				nodeStatus := v.Value.GetStatus()
				if nodeStatus == node.NodeStatusInit || nodeStatus == node.NodeStatusFailed {
					continue
				}
				allNodes = append(allNodes, v.Value)
			}

			allSubnet := []string{}
			for _, n := range allNodes {
				allSubnet = append(allSubnet, n.Subnet)
			}
			sort.Strings(allSubnet)
			sort.Strings(nodeSubnetCache)
			if utils.EqualStringSlice(allSubnet, nodeSubnetCache) {
				continue
			}

			if networkType == "hostgw" {
				if err := updateRoutes(allNodes); err == nil {
					log.Println("all nodes routes update successfully")
					nodeSubnetCache = allSubnet
				}
			}
		}
	}
}
