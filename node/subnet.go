package node

import (
	"log"

	"github.com/3th1nk/cidr"
	"github.com/JinlongWukong/CloudLab/config"
)

//subnet pool
var subnets = make([]string, 0)

//subnet range
var subnetRange = "192.168.0.0/16"

//translate subnet range to subnet pool
func init() {
	if config.Node.SubnetRange != "" {
		subnetRange = config.Node.SubnetRange
	}

	c, err := cidr.ParseCIDR(subnetRange)
	if err != nil {
		log.Printf("Parse given subnetRange failed %v", err)
		return
	}
	cs, err := c.SubNetting(cidr.SUBNETTING_METHOD_HOST_NUM, 256)
	if err != nil {
		log.Printf("Split given subnetRange failed %v", err)
		return
	}
	for _, c := range cs {
		subnets = append(subnets, c.CIDR())
	}
}

//allocate a subnet to node
func allocateSubnet() string {
	used := map[string]struct{}{}
	for v := range NodeDB.Iter() {
		used[v.Value.Subnet] = struct{}{}
	}

	for _, s := range subnets {
		if _, found := used[s]; !found {
			return s
		}
	}
	return ""
}
