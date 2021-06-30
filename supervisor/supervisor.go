package supervisor

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JinlongWukong/DevLab/config"
	"github.com/JinlongWukong/DevLab/db"
	"github.com/JinlongWukong/DevLab/deployer"
	"github.com/JinlongWukong/DevLab/manager"
	"github.com/JinlongWukong/DevLab/node"
	"github.com/JinlongWukong/DevLab/utils"
)

var nodeCheckInterval = "180s"
var nodeLimitCPU = 0.8
var nodeMinimumMem = 2048
var nodeLimitDisk = 80
var enable = "true"

type Supervisor struct {
}

var _ manager.Manager = Supervisor{}

func init() {
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
	if config.Supervisor.Enable != "" {
		enable = config.Supervisor.Enable
	}
}

func (s Supervisor) Control(ctx context.Context, wg *sync.WaitGroup) {

	log.Println("Supervisor manager started")
	defer func() {
		log.Println("Supervisor manager exited")
		wg.Done()
	}()

	if enable == "true" {
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
					if nodeStatus == node.NodeStatusInit || nodeStatus == node.NodeStatusInstallFailed {
						continue
					}

					select {
					case <-ctx.Done():
						return
					default:
						query := map[string]string{
							"Ip":   n.IpAddress,
							"Pass": n.Passwd,
							"User": n.UserName,
							"Role": string(n.Role),
						}
						log.Printf("Remote http call to check node usage %v", n.Name)
						var nodeCondition node.NodeCondition
						url := deployer.GetDeployerBaseUrl() + "/host"
						err, reponse_data := utils.HttpGetJsonData(url, query)
						if err != nil {
							log.Printf("Remote http call to check node %v usage failed with error -> %v", n.Name, err)
							if strings.Contains(err.Error(), "unexpected status-code returned") {
								//if err occur, set node as unhealth status
								n.SetStatus(node.NodeStatusUnhealth)
								db.NotifyToSave()
							}
						} else {
							log.Printf("Remote http call to check node %v successfully", n.Name)
							if err := json.Unmarshal(reponse_data, &nodeCondition); err != nil {
								log.Printf("Parse node %v response data failed -> %v", n.Name, err)
								break
							}
							diskUsage, _ := strconv.Atoi(strings.Split(nodeCondition.DiskUsage, "%")[0])
							//If at least one of below conditions not satisfied, means overload
							if nodeCondition.CpuLoad > float64(n.CPU)*nodeLimitCPU ||
								nodeCondition.MemAvail < nodeMinimumMem ||
								diskUsage > nodeLimitDisk {
								n.SetStatus(node.NodeStatusOverload)
							} else if nodeCondition.Engine != 0 {
								log.Printf("node %v engine %v is down", n.Name, n.Role)
								n.SetStatus(node.NodeStatusUnhealth)
							} else {
								n.SetStatus(node.NodeStatusReady)
							}
							db.NotifyToSave()
						}
					}
				}
			}
		}
	}
}
