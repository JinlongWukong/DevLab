package supervisor

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/deployer"
	"github.com/JinlongWukong/CloudLab/manager"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/utils"
)

var nodeCheckInterval = "10s"
var nodeLimitCPU = 0.8
var nodeMinimumMem = 2048
var nodeLimitDisk = 80

type Supervisor struct {
}

var _ manager.Manager = Supervisor{}

func initialize() {
	if config.Supervisor.NodeCheckInterval != "" {
		nodeCheckInterval = config.Supervisor.NodeCheckInterval
	}
	if config.Supervisor.NodeLimitCPU != 0 {
		nodeLimitCPU = config.Supervisor.NodeLimitCPU
	}
	if config.Supervisor.NodeMinimumMem != 0 {
		nodeMinimumMem = config.Supervisor.NodeMinimumMem
	}
	if config.Supervisor.NodeLimitDisk != 0 {
		nodeLimitDisk = config.Supervisor.NodeLimitDisk
	}
}

func (s Supervisor) Control(ctx context.Context, wg *sync.WaitGroup) {

	defer func() {
		log.Println("Supervisor manager exited")
		wg.Done()
	}()

	initialize()

	interval, err := time.ParseDuration(nodeCheckInterval)
	if err != nil {
		log.Println(err)
		return
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			//Copy all nodes
			allNodes := []*node.Node{}
			for v := range node.NodeDB.Iter() {
				allNodes = append(allNodes, v.Value)
			}

			for _, n := range allNodes {
				//if node status not installed or failed, skip
				nodeStatus := n.GetStatus()
				if nodeStatus == node.NodeStatusInit || nodeStatus == node.NodeStatusFailed {
					break
				}

				select {
				case <-ctx.Done():
					return
				default:
					query := map[string]string{
						"Ip":   n.IpAddress,
						"Pass": n.Passwd,
						"User": n.UserName,
					}
					log.Printf("Remote http call to check node usage %v", n.Name)
					var nodeUsage node.NodeUsage
					url := deployer.GetDeployerBaseUrl() + "/host"
					err, reponse_data := utils.HttpGetJsonData(url, query)
					if err != nil {
						log.Printf("Remote http call to check node %v usage failed with error -> %v", n.Name, err)
						//if err occur, set node as unhealth status
						n.SetStatus(node.NodeStatusUnhealth)
						db.NotifyToDB("node", n.Name, "update")
						break
					} else {
						log.Printf("Remote http call to check node %v successfully", n.Name)
						if err := json.Unmarshal(reponse_data, &nodeUsage); err != nil {
							log.Printf("Parse node %v response data failed -> %v", n.Name, err)
							break
						}
						diskUsage, _ := strconv.Atoi(strings.Split(nodeUsage.DiskUsage, "%")[0])
						//If at least one of below conditions not satisfied, means overload
						if nodeUsage.CpuLoad > float64(n.CPU)*nodeLimitCPU ||
							nodeUsage.MemAvail < nodeMinimumMem ||
							diskUsage > nodeLimitDisk {
							n.SetStatus(node.NodeStatusOverload)
						} else {
							n.SetStatus(node.NodeStatusReady)
						}
						db.NotifyToDB("node", n.Name, "update")
					}
				}
			}
		}
	}
}
