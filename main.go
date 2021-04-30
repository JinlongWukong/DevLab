package main

import (
	"log"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/vm"
	"github.com/JinlongWukong/CloudLab/workflow"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var account_db = make(map[string]*account.Account)

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	r.LoadHTMLGlob("views/*")
	r.GET("/", indexHandler)

	//vm related api
	r.GET("/vm-request", vmRequestIndexHandler)
	r.GET("/vm-request/vm", vmRequestGetHandler)

	r.POST("/vm-request", vmRequestPostHandler)
	r.POST("/vm-request/vm", vmRequestPostActionHandler)

	r.Run(":8088")
}

// Head Page
func indexHandler(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

// VM reqeust head page
func vmRequestIndexHandler(c *gin.Context) {
	c.HTML(200, "vmRequest.html", nil)
}

func vmRequestGetHandler(c *gin.Context) {
	var g vm.VmRequestGetVm
	c.Bind(&g)

	myaccount, exists := account_db[g.Account]
	if exists == true {
		if g.Name == "" {
			c.JSON(200, myaccount.VM)
		} else {
			if vmInfo, err := myaccount.GetVmByName(g.Name); err == nil {
				c.JSON(200, vmInfo)
			} else {
				c.JSON(404, gin.H{"VM": err})
			}
		}
	} else {
		c.JSON(404, "Account not found")
	}
}

// VM reqeust POST handler, create VM
func vmRequestPostHandler(c *gin.Context) {
	var vmRequest vm.VmRequest
	c.Bind(&vmRequest)
	log.Println(vmRequest.Account, vmRequest.Type, vmRequest.Number, vmRequest.Duration)

	myaccount, exists := account_db[vmRequest.Account]
	if exists == false {
		myaccount = &account.Account{Name: vmRequest.Account, Role: "guest"}
		account_db[vmRequest.Account] = myaccount
	}

	if myaccount.StatusVm == "running" {
		c.JSON(202, "VM creation is ongoing, please try later")
		return
	}

	go func() {
		workflow.CreateVMs(myaccount, vmRequest)
	}()

	c.JSON(200, "VM creation request accepted")
}

// VM request POST action handler, start/stop/reboot/delete VM
// Return:
//     20x     -> success
//     40x/50x -> failed
func vmRequestPostActionHandler(c *gin.Context) {
	var vmRequestAction vm.VmRequestPostAction
	c.Bind(&vmRequestAction)
	log.Printf("Get VM action request: Account -> %v, VM -> %v, Action -> %v ", vmRequestAction.Account, vmRequestAction.Name, vmRequestAction.Action)

	myaccount, exists := account_db[vmRequestAction.Account]
	if exists == true {
		if vmRequestAction.Name == "" || vmRequestAction.Action == "" {
			c.JSON(400, "VM name or Action empty")
		}
		if myVM, err := myaccount.GetVmByName(vmRequestAction.Name); err == nil {
			var action_err error
			switch vmRequestAction.Action {
			case "start":
				action_err = vm.StartUpVirtualMachine(myVM)
			case "shutdown":
				action_err = vm.ShutDownVirtualMachine(myVM)
			case "reboot":
				action_err = vm.RebootVirtualMachine(myVM)
			case "delete":
				action_err = vm.DeleteVirtualMachine(myVM)
			default:
				c.JSON(400, "Action not support")
			}
			if action_err != nil {
				c.JSON(500, err)
			}
		} else {
			c.JSON(404, "VM not found")
		}
	} else {
		c.JSON(404, "Account not found")
	}

	c.JSON(202, "")
}
