package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/vm"
	"github.com/JinlongWukong/CloudLab/workflow"
)

// Head Page
func IndexHandler(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func AdminHandler(c *gin.Context) {
	c.HTML(200, "admin.html", nil)
}

// Todo Page
func ToDoHandler(c *gin.Context) {
	c.JSON(200, "not implemented yet")
}

// VM reqeust index page
func VmRequestIndexHandler(c *gin.Context) {
	c.HTML(200, "vmRequest.html", nil)
}

// Get VMs of specify account
// Args:
//   VM Name or empty(means get all vm)
// Return:
//   20x: success with VM info
//   40x: fail Account/VM not found
func VmRequestGetVmHandler(c *gin.Context) {
	var g vm.VmRequestGetVm
	c.Bind(&g)

	myaccount, exists := account.AccountDB.Get(g.Account)
	if exists == true {
		if g.Name == "" {
			// return all vm
			c.JSON(http.StatusOK, myaccount.VM)
		} else {
			if myVM, err := myaccount.GetVmByName(g.Name); err == nil {
				c.JSON(http.StatusOK, myVM)
			} else {
				c.JSON(http.StatusNotFound, "VM not found")
			}
		}
	} else {
		c.JSON(http.StatusNotFound, "Account not found")
	}
}

// VM reqeust POST handler, Create VM
func VmRequestCreateVmHandler(c *gin.Context) {
	var vmRequest vm.VmRequest
	c.Bind(&vmRequest)
	log.Println(vmRequest.Account, vmRequest.Type, vmRequest.Flavor, vmRequest.Number, vmRequest.Duration)

	if vmRequest.Account == "" ||
		vmRequest.Type == "" ||
		vmRequest.Flavor == "" ||
		vmRequest.Number < 1 {
		c.JSON(http.StatusBadRequest, "input parameters error")
		return
	}

	myaccount, exists := account.AccountDB.Get(vmRequest.Account)
	if exists == false {
		// Acount not existed, add new
		myaccount = &account.Account{Name: vmRequest.Account, Role: "guest"}
		account.AccountDB.Set(vmRequest.Account, myaccount)
	}

	if myaccount.StatusVm == "running" {
		c.JSON(http.StatusAccepted, "VM creation is ongoing, please try later")
		return
	}

	if err := workflow.CreateVMs(myaccount, vmRequest); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "VM creation request accepted")
}

// VM request VM action handler, start/stop/reboot/delete VM
// Return:
//     20x     -> success
//     40x/50x -> failed
func VmRequestVmActionHandler(c *gin.Context) {
	var vmRequestAction vm.VmRequestPostAction
	c.ShouldBind(&vmRequestAction)
	log.Printf("Get VM action request: Account -> %v, VM -> %v, Action -> %v ", vmRequestAction.Account, vmRequestAction.Name, vmRequestAction.Action)

	myaccount, exists := account.AccountDB.Get(vmRequestAction.Account)
	if exists == true {
		if vmRequestAction.Name == "" || vmRequestAction.Action == "" {
			c.JSON(http.StatusBadRequest, "VM name or Action empty")
		}
		if myVM, err := myaccount.GetVmByName(vmRequestAction.Name); err == nil {
			var action_err error
			switch vmRequestAction.Action {
			case "start", "shutdown", "reboot", "delete":
				action_err = workflow.ActionVM(myaccount, myVM, vmRequestAction.Action)
			default:
				c.JSON(http.StatusBadRequest, "Action not support")
				return
			}
			if action_err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			c.JSON(http.StatusNotFound, "VM not found")
			return
		}
	} else {
		c.JSON(http.StatusNotFound, "Account not found")
		return
	}

	c.JSON(http.StatusNoContent, "")
	return
}

// VM request port expose handler,
// Return:
//     20x     -> success
//     40x/50x -> failed
func VmRequestVmPortExposeHandler(c *gin.Context) {
	var vmRequestPortExpose vm.VmRequestPortExpose
	c.ShouldBind(&vmRequestPortExpose)
	log.Printf("Get VM port expose request: Account -> %v, VM -> %v, Port -> %v, Protocol -> %v ", vmRequestPortExpose.Account, vmRequestPortExpose.Name,
		vmRequestPortExpose.Port, vmRequestPortExpose.Protocol)

	myaccount, exists := account.AccountDB.Get(vmRequestPortExpose.Account)
	if exists == true {
		if myVM, err := myaccount.GetVmByName(vmRequestPortExpose.Name); err == nil {
			var action_err error
			action_err = workflow.ExposePort(myaccount, myVM, vmRequestPortExpose.Port, vmRequestPortExpose.Protocol)
			if action_err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			c.JSON(http.StatusNotFound, "VM not found")
			return
		}
	} else {
		c.JSON(http.StatusNotFound, "Account not found")
		return
	}

	c.JSON(http.StatusNoContent, "")
	return
}

// Get nodes info
// Request:
//   node name or empty(means get all nodes)
// Return:
//   200: success -> with Node info
//   404: fail -> Node not found
func NodeRequestGetNodeHandler(c *gin.Context) {
	x, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Printf("%s", string(x))
	var r node.NodeRequest
	c.ShouldBind(&r)

	if r.Name == "" {
		log.Println("Receive node request -> get all nodes info")
		allNodesDetails := []*node.Node{}
		for v := range node.NodeDB.Iter() {
			allNodesDetails = append(allNodesDetails, v.Value)
		}
		c.JSON(http.StatusOK, allNodesDetails)
	} else {
		log.Printf("Receive node request -> get node %v info", r.Name)
		mynode, exists := node.NodeDB.Get(r.Name)
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
func NodeRequestActionNodeHandler(c *gin.Context) {

	var nodeRequest node.NodeRequest
	c.ShouldBind(&nodeRequest)
	log.Printf("Node request coming, node name -> %v action -> %v", nodeRequest.Name, nodeRequest.Action)

	switch nodeRequest.Action {
	case node.NodeActionAdd:
		_, exists := node.NodeDB.Get(nodeRequest.Name)
		if exists == true {
			log.Printf("Node %v already existed", nodeRequest.Name)
		} else {
			go func() {
				workflow.AddNode(nodeRequest)
			}()
		}
		c.JSON(http.StatusOK, "")
	case node.NodeActionRemove, node.NodeActionReboot, node.NodeActionEnable, node.NodeActionDisable:
		_, exists := node.NodeDB.Get(nodeRequest.Name)
		if exists == true {
			if err := workflow.ActionNode(nodeRequest); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			} else {
				c.JSON(http.StatusOK, "")
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
}

//Will return total task numbers
func WorkflowTaskHandler(c *gin.Context) {

	taskNumber := workflow.GetTaskCount()
	c.Writer.WriteString(fmt.Sprintf("taskNumber %v", taskNumber))
}
