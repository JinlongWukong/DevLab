package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/vm"
	"github.com/JinlongWukong/CloudLab/workflow"
)

func main() {

	//Start db control loop
	db.Manager()

	r := gin.Default()
	r.Use(cors.Default())

	r.LoadHTMLGlob("views/*")
	r.GET("/", indexHandler)

	//vm related api
	r.GET("/vm-request", vmRequestIndexHandler)
	r.GET("/vm-request/vm", vmRequestGetVmHandler)

	//TODO api
	r.GET("/k8s-request", toDoHandler)
	r.GET("/container-request", toDoHandler)

	r.Run(":8088")
}

// Head Page
func indexHandler(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

// Todo Page
func toDoHandler(c *gin.Context) {
	c.JSON(200, "not implemented yet")
}

// VM reqeust index page
func vmRequestIndexHandler(c *gin.Context) {
	c.HTML(200, "vmRequest.html", nil)
}

// Get VMs of specify account
// Args:
//   VM Name or empty(means get all vm)
// Return:
//   20x: success with VM info
//   40x: fail Account/VM not found
func vmRequestGetVmHandler(c *gin.Context) {
	var g vm.VmRequestGetVm
	c.Bind(&g)

	myaccount, exists := account.Account_db[g.Account]
	if exists == true {
		if g.Name == "" {
			// return all vm
			c.JSON(200, myaccount.VM)
		} else {
			if myVM, err := myaccount.GetVmByName(g.Name); err == nil {
				c.JSON(200, myVM)
			} else {
				c.JSON(404, "VM not found")
			}
		}
	} else {
		c.JSON(404, "Account not found")
	}
}

// VM reqeust POST handler, Create VM
func vmRequestCreateVmHandler(c *gin.Context) {
	var vmRequest vm.VmRequest
	c.Bind(&vmRequest)
	log.Println(vmRequest.Account, vmRequest.Type, vmRequest.Number, vmRequest.Duration)

	myaccount, exists := account.Account_db[vmRequest.Account]
	if exists == false {
		// Acount not existed, add new
		myaccount = &account.Account{Name: vmRequest.Account, Role: "guest"}
		account.Account_db[vmRequest.Account] = myaccount
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

// VM request VM action handler, start/stop/reboot/delete VM
// Return:
//     20x     -> success
//     40x/50x -> failed
func vmRequestVmActionHandler(c *gin.Context) {
	var vmRequestAction vm.VmRequestPostAction
	c.Bind(&vmRequestAction)
	log.Printf("Get VM action request: Account -> %v, VM -> %v, Action -> %v ", vmRequestAction.Account, vmRequestAction.Name, vmRequestAction.Action)

	myaccount, exists := account.Account_db[vmRequestAction.Account]
	if exists == true {
		if vmRequestAction.Name == "" || vmRequestAction.Action == "" {
			c.JSON(400, "VM name or Action empty")
		}
		if myVM, err := myaccount.GetVmByName(vmRequestAction.Name); err == nil {
			var action_err error
			switch vmRequestAction.Action {
			case "start", "shutdown", "reboot", "delete":
				action_err = workflow.ActionVM(myaccount, myVM, vmRequestAction.Action)
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
