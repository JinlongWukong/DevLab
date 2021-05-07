package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/node"
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
	r.GET("/admin", adminHandler)

	//vm related api
	r.GET("/vm-request", vmRequestIndexHandler)
	r.GET("/vm-request/vm", vmRequestGetVmHandler)

	r.POST("/vm-request", vmRequestCreateVmHandler)
	r.POST("/vm-request/vm", vmRequestVmActionHandler)

	//node related api
	r.GET("/node-request", nodeRequestGetNodeHandler)
	r.POST("/node-request", nodeRequestActionNodeHandler)

	//TODO api
	r.GET("/k8s-request", toDoHandler)
	r.GET("/container-request", toDoHandler)

	r.Run(":8088")
}

// Head Page
func indexHandler(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func adminHandler(c *gin.Context) {
	c.HTML(200, "admin.html", nil)
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
	c.ShouldBind(&vmRequestAction)
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

// Get nodes info
// Args:
//   node Name or empty(means get all nodes)
// Return:
//   20x: success with Node info
//   404: fail Node not found
func nodeRequestGetNodeHandler(c *gin.Context) {

	var g node.NodeRequestGetNode
	c.ShouldBind(&g)

	if g.Name == "" {
		log.Println("Received get all nodes info request")
		c.JSON(200, node.Node_db)
	} else {
		log.Printf("Received get node %v info request", g.Name)
		mynode, exists := node.Node_db[g.Name]
		if exists == true {
			c.JSON(200, mynode)
		} else {
			c.JSON(404, "Node not found")
		}
	}
}

// Node request action handler, add/remove/reboot node
// If action is add node -> async call, otherwise -> sync call
func nodeRequestActionNodeHandler(c *gin.Context) {

	var nodeRequest node.NodeRequest
	c.ShouldBind(&nodeRequest)
	log.Printf("Node request coming, %v %v", nodeRequest.Name, nodeRequest.Action)

	switch nodeRequest.Action {
	case "add":
		_, exists := node.Node_db[nodeRequest.Name]
		if exists == true {
			c.JSON(400, "Node already existed")
		} else {
			go func() {
				workflow.AddNode(nodeRequest)
			}()
		}
	case "remove", "reboot":
		_, exists := node.Node_db[nodeRequest.Name]
		if exists == true {
			if err := workflow.ActionNode(nodeRequest); err != nil {
				c.JSON(500, err)
			}
		} else {
			c.JSON(400, "Node not existed")
		}
	default:
		c.JSON(400, "Action not support")
	}

	c.JSON(200, "success")
}
