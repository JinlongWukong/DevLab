package api

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JinlongWukong/DevLab/account"
	"github.com/JinlongWukong/DevLab/auth"
	"github.com/JinlongWukong/DevLab/k8s"
	"github.com/JinlongWukong/DevLab/node"
	"github.com/JinlongWukong/DevLab/saas"
	"github.com/JinlongWukong/DevLab/vm"
	"github.com/JinlongWukong/DevLab/workflow"
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

//K8S request index page
func K8sRequestIndexHandler(c *gin.Context) {
	c.HTML(200, "k8sRequest.html", nil)
}

//SaaS request index page
func SoftwareIndexHandler(c *gin.Context) {
	c.HTML(200, "SaaSRequest.html", nil)
}

// Get VMs of specify account
// Args:
//   VM Name or empty(means get all vm)
// Return:
//   20x: success with VM info
//   40x: fail Account/VM not found
func VmRequestGetVmHandler(c *gin.Context) {
	var g vm.VmRequestGetVm
	if err := c.Bind(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	if err := c.Bind(&vmRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	if _, err := workflow.CreateVMs(myaccount, vmRequest); err != nil {
		c.JSON(http.StatusInternalServerError, "")
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
	if err := c.Bind(&vmRequestAction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
			case "extend":
				action_err = workflow.ExtendVMLifetime(myVM, 24*time.Hour)
			default:
				c.JSON(http.StatusBadRequest, "Action not support")
				return
			}
			if action_err != nil {
				c.JSON(http.StatusInternalServerError, action_err.Error())
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
	if err := c.Bind(&vmRequestPortExpose); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Get VM port expose request: Account -> %v, VM -> %v, Port -> %v, Protocol -> %v ", vmRequestPortExpose.Account, vmRequestPortExpose.Name,
		vmRequestPortExpose.Port, vmRequestPortExpose.Protocol)

	myaccount, exists := account.AccountDB.Get(vmRequestPortExpose.Account)
	if exists == true {
		if myVM, err := myaccount.GetVmByName(vmRequestPortExpose.Name); err == nil {
			var action_err error
			action_err = workflow.ExposePort(myaccount, myVM, vmRequestPortExpose.Port, vmRequestPortExpose.Protocol)
			if action_err != nil {
				c.JSON(http.StatusInternalServerError, action_err.Error())
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
	var r node.NodeRequest
	if err := c.Bind(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	if err := c.Bind(&nodeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

// K8s request POST handler, Create K8S
func K8sRequestCreateHandler(c *gin.Context) {
	var k8sRequest k8s.K8sRequest
	if err := c.Bind(&k8sRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println(k8sRequest.Account, k8sRequest.Version,
		k8sRequest.NumOfContronller, k8sRequest.NumOfWorker)

	if k8sRequest.Account == "" ||
		k8sRequest.Version == "" ||
		k8sRequest.NumOfWorker < 1 ||
		k8sRequest.NumOfContronller < 1 {
		c.JSON(http.StatusBadRequest, "input parameters error")
		return
	}

	myaccount, exists := account.AccountDB.Get(k8sRequest.Account)
	if exists == false {
		// Acount not existed, add new
		myaccount = &account.Account{Name: k8sRequest.Account, Role: "guest"}
		account.AccountDB.Set(k8sRequest.Account, myaccount)
	}

	if err := workflow.CreateK8S(myaccount, k8sRequest); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "K8S creation request accepted")
}

// K8s request Delete handler, remove K8S
func K8sRequestDeleteHandler(c *gin.Context) {
	var g k8s.K8sRequestAction
	if err := c.Bind(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println(g)

	if g.Name == "" {
		err_msg := "k8s name not specify"
		log.Println(err_msg)
		c.JSON(http.StatusBadRequest, err_msg)
	} else {
		if err := workflow.DeleteK8S(g); err == nil {
			c.JSON(http.StatusOK, "")
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err.Error())
		}
	}
}

// Get k8s of specify account
// Args:
//   k8s Name or empty(means get all k8s)
// Return:
//   20x: success with k8s info
//   40x: fail Account/k8s not found
func K8sRequestGetHandler(c *gin.Context) {
	var g k8s.K8sRequestAction
	if err := c.Bind(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	myaccount, exists := account.AccountDB.Get(g.Account)
	if exists == true {
		if g.Name == "" {
			// return all k8s
			c.JSON(http.StatusOK, myaccount.K8S)
		} else {
			if myK8s, err := myaccount.GetK8sByName(g.Name); err == nil {
				c.JSON(http.StatusOK, myK8s)
			} else {
				c.JSON(http.StatusNotFound, "k8s not found")
			}
		}
	} else {
		c.JSON(http.StatusNotFound, "Account not found")
	}
}

// Software part

// Software request create handler
// Args:
//   software request
// Return:
//     20x     -> success
//     40x/50x -> failed
func SoftwareRequestCreateHandler(c *gin.Context) {
	var request saas.SoftwareRequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Get software creation request, %v,%v,%v", request.Account, request.Kind, request.Version)

	myaccount, exists := account.AccountDB.Get(request.Account)
	if exists == false {
		// Acount not existed, add new
		myaccount = &account.Account{Name: request.Account, Role: "guest"}
		account.AccountDB.Set(request.Account, myaccount)
	}

	if err := workflow.CreateSoftware(myaccount, request); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "Software creation request accepted")
}

// Software request action handler, start/stop/delete
// Args:
//   software Name
// Return:
//     20x     -> success
//     40x/50x -> failed
func SoftwareRequestActionHandler(c *gin.Context) {
	var r saas.SoftwareRequestAction
	if err := c.Bind(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Get Software action request: Account -> %v, Software -> %v, Action -> %v ", r.Account, r.Name, r.Action)

	myAccount, exists := account.AccountDB.Get(r.Account)
	if exists == true {
		if r.Name == "" || r.Action == "" {
			c.JSON(http.StatusBadRequest, "Software name or Action empty")
		}
		var action_err error
		switch r.Action {
		case "start", "stop", "restart", "delete", "get":
			action_err = workflow.ActionSoftware(myAccount, r)
		default:
			c.JSON(http.StatusBadRequest, "Action not support")
			return
		}
		if action_err != nil {
			c.JSON(http.StatusInternalServerError, action_err.Error())
			return
		}

	} else {
		c.JSON(http.StatusNotFound, "Account not found")
		return
	}

	c.JSON(http.StatusNoContent, "")
	return
}

// Get software of specify account
// Args:
//   software Name or empty(means get all software)
// Return:
//   20x: success with software info
//   40x: fail Account/software not found
func SoftwareRequestGetHandler(c *gin.Context) {
	var g saas.SoftwareRequestGetInfo
	if err := c.Bind(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	myaccount, exists := account.AccountDB.Get(g.Account)
	if exists == true {
		if g.Name == "" {
			// return all software
			c.JSON(http.StatusOK, myaccount.Software)
		} else {
			if mySoftware, err := myaccount.GetSoftwareByName(g.Name); err == nil {
				c.JSON(http.StatusOK, mySoftware)
			} else {
				c.JSON(http.StatusNotFound, "software not found")
			}
		}
	} else {
		c.JSON(http.StatusNotFound, "Account not found")
	}
}

//Will return total task numbers
func WorkflowTaskHandler(c *gin.Context) {

	taskNumber := workflow.GetTaskCount()
	c.Writer.WriteString(fmt.Sprintf("taskNumber %v", taskNumber))
}

//metrics handler
func metricsHandler(c *gin.Context) {

	c.Writer.WriteString(fmt.Sprintf("GoroutineNumber %v", runtime.NumGoroutine()))

}

// account one-time password generate handler
func oneTimePassGenHandler(c *gin.Context) {

	accountName := c.Query("account")

	if accountName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account name not give"})
		return
	}

	myaccount, exists := account.AccountDB.Get(accountName)
	if exists == true {
		myaccount.SetOneTimePass(true)
		c.JSON(http.StatusOK, nil)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
	}

	return
}

// account login handler
/*
{
 "access_token": "xxxxxxxxxxxxxxxxxxx",
 "token_type": "bearer",
 "expires_in": 86400
}
*/
func accountLogin(c *gin.Context) {
	var r auth.LoginInfo
	if err := c.Bind(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	myaccount, exists := account.AccountDB.Get(r.Account)
	if exists == true {
		if r.Password == myaccount.GetOneTimePass() {
			log.Printf("account %v login successfully", r.Account)
			//Clear one time password after a success login
			myaccount.SetOneTimePass(false)
			tokenInfo := auth.InvokeToken(r.Account)
			c.JSON(http.StatusOK, tokenInfo)
		} else {
			c.JSON(http.StatusUnauthorized, nil)
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
	}
}
