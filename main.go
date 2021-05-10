package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
// Request:
//   node name or empty(means get all nodes)
// Return:
//   200: success -> with Node info
//   404: fail -> Node not found
func nodeRequestGetNodeHandler(c *gin.Context) {
	x, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Printf("%s", string(x))
	var r node.NodeRequest
	c.ShouldBind(&r)

	if r.Name == "" {
		log.Println("Receive node request -> get all nodes info")
		allNodesDetails := []*node.Node{}
		for _, info := range node.Node_db {
			allNodesDetails = append(allNodesDetails, info)
		}
		c.JSON(http.StatusOK, allNodesDetails)
	} else {
		log.Printf("Receive node request -> get node %v info", r.Name)
		mynode, exists := node.Node_db[r.Name]
		if exists == true {
			c.JSON(http.StatusOK, mynode)
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Node not existed",
			})
		}
	}
}

// Node request action handler -> add/remove/reboot node
// If action is add node -> async call, otherwise -> sync call
func nodeRequestActionNodeHandler(c *gin.Context) {

	var nodeRequest node.NodeRequest
	c.ShouldBind(&nodeRequest)
	log.Printf("Node request coming, node name -> %v action -> %v", nodeRequest.Name, nodeRequest.Action)

	switch nodeRequest.Action {
	case node.NodeActionAdd:
		_, exists := node.Node_db[nodeRequest.Name]
		if exists == true {
			log.Printf("Node %v already existed", nodeRequest.Name)
		} else {
			go func() {
				workflow.AddNode(nodeRequest)
			}()
		}
	case node.NodeActionRemove, node.NodeActionReboot, node.NodeActionEnable, node.NodeActionDisable:
		_, exists := node.Node_db[nodeRequest.Name]
		if exists == true {
			if err := workflow.ActionNode(nodeRequest); err != nil {
				c.JSON(500, gin.H{
					"error": err,
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Node not existed",
			})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "action not support",
		})
	}

	c.JSON(http.StatusOK, "")
}
