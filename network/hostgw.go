package network

import (
	"encoding/json"
	"log"

	"github.com/JinlongWukong/DevLab/deployer"
	"github.com/JinlongWukong/DevLab/node"
	"github.com/JinlongWukong/DevLab/utils"
)

func updateRoutes(nodes []*node.Node) error {

	hosts := [][]string{}
	routes := []map[string]string{}

	for _, n := range nodes {
		login := []string{n.IpAddress, n.UserName, n.Passwd}
		route := map[string]string{"subnet": n.Subnet, "via": n.IpAddress}
		hosts = append(hosts, login)
		routes = append(routes, route)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"Hosts":  hosts,
		"Routes": routes,
		"Action": "route",
	})

	log.Println("Remote http call to update node route table")

	url := deployer.GetDeployerBaseUrl() + "/hosts"
	if err, _ := utils.HttpSendJsonData(url, "POST", payload); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
