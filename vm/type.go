package vm

import (
	"time"
)

var flavorDetails = map[string]map[string]int32{
	"small": {
		"cpu":    2,
		"memory": 2048,
		"disk":   30,
	},
	"middle": {
		"cpu":    4,
		"memory": 4096,
		"disk":   64,
	},
	"large": {
		"cpu":    6,
		"memory": 8192,
		"disk":   80,
	},
}

type VncInfo struct {
	Port string `json:"Port"`
	Pass string `json:"Passwd"`
}

type VirtualMachine struct {
	Name      string        `json:"Name"`
	CPU       int32         `json:"CPU"`
	Memory    int32         `json:"Mem"`
	Disk      int32         `json:"Disk"`
	IpAddress string        `json:"Address"`
	Status    string        `json:"Status"`
	Vnc       VncInfo       `json:"Vnc"`
	Type      string        `json:"Type"`
	Node      string        `json:"Node"`
	Lifetime  time.Duration `json:"LifeTime"`
}

type VmRequest struct {
	Account  string `form:"cecid"`
	Type     string `form:"os_type"`
	Flavor   string `form:"os_flavor"`
	Number   int32  `form:"os_numbers"`
	Duration int    `form:"os_duration"`
}

type VmRequestGetVm struct {
	Account string `form:"cecid"`
	Name    string `form:"vm_name"`
}

type VmRequestPostAction struct {
	Account string `form:"cecid"`
	Name    string `form:"vm_name"`
	Action  string `form:"vm_action"`
}

type VmLiveStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Address string `json:"address"`
	VncPort string `json:"vnc_port"`
}
