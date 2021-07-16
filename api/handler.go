package api

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/JinlongWukong/DevLab/account"
	"github.com/JinlongWukong/DevLab/auth"
	"github.com/JinlongWukong/DevLab/db"
	"github.com/JinlongWukong/DevLab/k8s"
	"github.com/JinlongWukong/DevLab/node"
	"github.com/JinlongWukong/DevLab/saas"
	"github.com/JinlongWukong/DevLab/terminal"
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

// Get VMs
// Return:
//   200: success with all VM info
//   404: fail Account not found
func VmRequestGetAllHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	log.Printf("Recevie vm request get all vm: %v", ac)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		c.JSON(http.StatusOK, myaccount.VM)
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
	}

}

// Get VM by name
// Return:
//   200: success with VM info
//   404: fail Account/VM not found
func VmRequestGetByNameHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	name := c.Param("name")
	log.Printf("Recevie vm request get by name: %v, %v", ac, name)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		if myVM, err := myaccount.GetVmByName(name); err == nil {
			c.JSON(http.StatusOK, myVM)
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "VM not found",
			})
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
	}

}

// Create VM
// This is async call
// Return:
//   200: success with VM info
//   404: fail Account/VM not found
func VmRequestCreateHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	var vmRequest vm.VmRequest
	if err := c.Bind(&vmRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	log.Printf("Recevie vm request to create vm: %v, %v, %v, %v, %v", ac, vmRequest.Type, vmRequest.Flavor, vmRequest.Number, vmRequest.Duration)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		if _, err := workflow.CreateVMs(myaccount, vmRequest); err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
		} else {
			c.JSON(http.StatusOK, "VM creation request accepted")
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Account not found",
		})
	}

}

// VM action, start/stop/reboot/delete
// Return:
//     204     -> success
//     40x/50x -> failed
func VmRequestActionHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	name := c.Param("name")
	action := c.Param("action")
	log.Printf("Receive VM action request: %v, %v, %v ", ac, name, action)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		if myVM, err := myaccount.GetVmByName(name); err == nil {
			var action_err error
			switch action {
			case "start", "shutdown", "reboot", "delete":
				action_err = workflow.ActionVM(myaccount, myVM, action)
			case "extend":
				action_err = workflow.ExtendVMLifetime(myVM, 24*time.Hour)
			default:
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Action not support",
				})
				return
			}
			if action_err != nil {
				c.JSON(http.StatusInternalServerError, action_err.Error())
				return
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "VM not found",
			})
			return
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)

}

// VM port expose,
// Return:
//     20x     -> success
//     40x/50x -> failed
func VmRequestPortExposeHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	name := c.Param("name")
	var vmRequestPortExpose vm.VmRequestPortExpose
	if err := c.Bind(&vmRequestPortExpose); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Receive VM port expose request: %v, %v, %v, %v ", ac, name,
		vmRequestPortExpose.Port, vmRequestPortExpose.Protocol)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		if myVM, err := myaccount.GetVmByName(name); err == nil {
			var action_err error
			action_err = workflow.ExposePort(myaccount, myVM, vmRequestPortExpose.Port, vmRequestPortExpose.Protocol)
			if action_err != nil {
				c.JSON(http.StatusInternalServerError, action_err.Error())
				return
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "VM not found",
			})
			return
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)

}

// VM inter-connect bet websoket and ssh channel
func VmRequestWebConsole(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Error(err)
		return
	}

	tokenString := c.Query("token")
	err, ac := auth.ValidateToken(tokenString)
	if err != nil {
		c.Error(err)
		return
	}
	vmName := c.Param("name")

	log.Printf("Receive VM web console request: %v,%v", ac, vmName)

	if ac, exists := account.AccountDB.Get(ac); exists == true {
		if myVm, err := ac.GetVmByName(vmName); err == nil {
			port, _ := strconv.Atoi(strings.Split(myVm.PortMap[22], ":")[0])
			host := node.GetNodeByName(myVm.Node)
			if host == nil {
				return
			}
			webTerminal := terminal.NewSSHTerminal("root", myVm.RootPass, host.IpAddress, uint16(port))
			err = webTerminal.Connect()
			if err != nil {
				conn.WriteMessage(1, []byte(err.Error()))
				conn.Close()
				return
			}
			webTerminal.NewShellTerminal()
			webTerminal.Ws2ssh(conn)
		}
	}
}

// Get all nodes information
// Return:
//   200: success -> all node infor
//   404: fail -> Node not found
func NodeRequestGetAllHandler(c *gin.Context) {

	log.Println("Receive node request to get all nodes info")
	allNodesDetails := []*node.Node{}
	for v := range node.NodeDB.Iter() {
		allNodesDetails = append(allNodesDetails, v.Value)
	}
	c.JSON(http.StatusOK, allNodesDetails)

}

// Get node information by name
// Return:
//   200: success -> node infor
//   404: fail -> node not found
func NodeRequestGetByNameHandler(c *gin.Context) {

	name := c.Param("name")
	log.Printf("Receive node request to get node %v info", name)
	if n, exists := node.NodeDB.Get(name); exists {
		c.JSON(http.StatusOK, n)
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Node not found",
		})
	}

}

// Install and add node
// This is async call
// Return
//     20x -> success
//     400 -> bad request
func NodeRequestCreateHandler(c *gin.Context) {

	var nodeRequest node.NodeRequest
	if err := c.Bind(&nodeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Receive node request, install and add node %v, %v, %v", nodeRequest.Name, nodeRequest.IpAddress, nodeRequest.Role)

	_, exists := node.NodeDB.Get(nodeRequest.Name)
	if exists == true {
		log.Printf("Node %v already existed, nothing to do", nodeRequest.Name)
		c.JSON(http.StatusNoContent, gin.H{
			"warn": "node already existed",
		})
	} else {
		go func() {
			workflow.AddNode(nodeRequest)
		}()
		c.JSON(http.StatusOK, nil)
	}

}

// Node request action handler -> remove/reboot/enable/disable node
// This is sync call
//   200: success
//   400: fail -> bad request
//   404: fail -> node not found
//   500: fail -> internal error
func NodeRequestActionHandler(c *gin.Context) {

	name := c.Param("name")
	action := c.Param("action")
	log.Printf("Receive node request, node name -> %v action -> %v", name, action)

	switch node.NodeAction(action) {
	case node.NodeActionRemove, node.NodeActionReboot, node.NodeActionEnable, node.NodeActionDisable:
		_, exists := node.NodeDB.Get(name)
		if exists == true {
			if err := workflow.ActionNode(name, node.NodeAction(action)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			} else {
				c.JSON(http.StatusOK, nil)
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "node not existed",
			})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "action not support",
		})
	}

}

// Create K8S
// Return:
//   200: success
//   404: fail -> account not found
//   500: fail -> workflow k8s create failed
func K8sRequestCreateHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	var k8sRequest k8s.K8sRequest
	if err := c.Bind(&k8sRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Recevie k8s create request: %v, %v, %v, %v", ac, k8sRequest.Version,
		k8sRequest.NumOfContronller, k8sRequest.NumOfWorker)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == false {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "account not found",
		})
		return
	}

	if err := workflow.CreateK8S(myaccount, k8sRequest); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "K8S creation request accepted")

}

// Delete k8s cluster by name
// Return:
//   200: success
//   404: fail -> account not found
//   500: fail -> workflow k8s delete failed
func K8sRequestDeleteHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	name := c.Param("name")
	log.Printf("Recevie k8s delete request: %v, %v", ac, name)

	myAccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		if err := workflow.DeleteK8S(myAccount, name); err == nil {
			c.JSON(http.StatusOK, nil)
		} else {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "account not found",
		})
	}

}

// Get all k8s
// Return:
//   200: success with k8s info
//   40x: fail Account/k8s not found
func K8sRequestGetAllHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	log.Printf("Recevie k8s get all request: %v", ac)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		c.JSON(http.StatusOK, myaccount.K8S)
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
	}

}

// Get all k8s by name
// Return:
//   200: success with k8s info
//   40x: fail Account/k8s not found
func K8sRequestGetByNameHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	name := c.Param("name")
	log.Printf("Recevie k8s get request: %v, %v", ac, name)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		if myk8S, err := myaccount.GetK8sByName(name); err == nil {
			c.JSON(http.StatusOK, myk8S)
		} else {
			c.JSON(http.StatusNotFound, "k8s not found")
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "account not found",
		})
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

	ac := c.GetHeader("account")
	var request saas.SoftwareRequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Recevie software creation request, %v, %v, %v", ac, request.Kind, request.Version)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == false {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "account not found"})
		return
	}

	if err := workflow.CreateSoftware(myaccount, request); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, "Software creation request accepted")
}

// Software request action handler, start/stop/delete/restart/refresh
// Return:
//     20x     -> success
//     40x/50x -> failed
func SoftwareRequestActionHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	name := c.Param("name")
	action := c.Param("action")
	log.Printf("Recevie Software action request: %v, %v, %v ", ac, name, action)

	myAccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		var action_err error
		switch saas.SoftwareAction(action) {
		case saas.SoftwareActionGet, saas.SoftwareActionStart, saas.SoftwareActionDelete,
			saas.SoftwareActionStop, saas.SoftwareActionRestart:
			action_err = workflow.ActionSoftware(myAccount, name, saas.SoftwareAction(action))
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "action not support",
			})
		}
		if action_err != nil {
			c.JSON(http.StatusInternalServerError, action_err.Error())
		} else {
			c.JSON(http.StatusNoContent, nil)
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "account not found",
		})
	}

}

// Get all software info
// Return:
//   200: success with software info
//   404: fail Account not found
func SoftwareRequestGetAllHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	log.Printf("Recevie Software get all request: %v", ac)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		c.JSON(http.StatusOK, myaccount.Software)
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
	}

}

// Container inter-connect bet websoket and docker api
func ContainerRequestWebConsole(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Error(err)
		return
	}

	tokenString := c.Query("token")
	err, ac := auth.ValidateToken(tokenString)
	if err != nil {
		c.Error(err)
		return
	}
	containerName := c.Param("name")

	log.Printf("Receive saas web console request: %v,%v", ac, containerName)

	if ac, exists := account.AccountDB.Get(ac); exists == true {
		if mySoftware, err := ac.GetSoftwareByName(containerName); err == nil {
			host := node.GetNodeByName(mySoftware.Node)
			if host == nil {
				return
			}
			/* insecure way, degraded
			containerTerminal := terminal.NewContainerTerminal(host.IpAddress, mySoftware.Name, 2375)
			if err := containerTerminal.Create(); err != nil {
				return
			}
			containerTerminal.Start(conn)*/
			webTerminal := terminal.NewSSHTerminal("root", host.Passwd, host.IpAddress, 22)
			err = webTerminal.Connect()
			if err != nil {
				conn.WriteMessage(1, []byte(err.Error()))
				conn.Close()
				return
			}
			cmd := "docker exec -it " + mySoftware.Name + " sh" + "\n"
			webTerminal.NewInteractiveCmdTerminal(conn, cmd)
		}
	}
}

// Get software info
// Return:
//   200: success with software info
//   404: fail -> Account/Software not found
func SoftwareRequestGetByNameHandler(c *gin.Context) {

	ac := c.GetHeader("account")
	name := c.Param("name")
	log.Printf("Recevie Software get request: %v, %v", ac, name)

	myaccount, exists := account.AccountDB.Get(ac)
	if exists == true {
		if mySoftware, err := myaccount.GetSoftwareByName(name); err == nil {
			c.JSON(http.StatusOK, mySoftware)
		} else {
			c.JSON(http.StatusNotFound, "software not found")
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "account not found",
		})
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

// generate one-time password
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

// use account/one-time-password to login for fetching a token
/*
{
 "access_token": "xxxxxxxxxxxxxxxxxxx",
 "token_type": "bearer",
 "expires_in": 86400
}
*/
func accountLoginHandler(c *gin.Context) {
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

//Get all account information
func AccountRequestGetAllHandler(c *gin.Context) {

	accountSlice := []map[string]string{}
	for ac := range account.AccountDB.Iter() {
		account := map[string]string{
			"name":     ac.Value.Name,
			"role":     string(ac.Value.Role),
			"contract": ac.Value.Contract,
		}
		accountSlice = append(accountSlice, account)
	}

	c.JSON(http.StatusOK, accountSlice)
}

//Get account information by name
func AccountRequestGetByNameHandler(c *gin.Context) {

	name := c.Param("name")
	if ac, exists := account.AccountDB.Get(name); exists {
		c.JSON(http.StatusOK, map[string]string{
			"name":     ac.Name,
			"role":     string(ac.Role),
			"contract": ac.Contract,
		})
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}

//Delete account
func AccountRequestDelByNameHandler(c *gin.Context) {

	name := c.Param("name")
	if ac, exists := account.AccountDB.Get(name); exists {
		if ac.GetNumbersOfVm() > 0 ||
			ac.GetNumbersOfK8s() > 0 ||
			ac.GetNumbersOfSoftware() > 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "account still have resouces created"})
		} else {
			account.AccountDB.Del(name)
			db.NotifyToSave()
			c.JSON(http.StatusNoContent, nil)
		}
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}

//Create a new account
//input params: name and role
func AccountRequestCreateHandler(c *gin.Context) {
	var r account.AccountRequest
	if err := c.Bind(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := account.AccountDB.Add(r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		db.NotifyToSave()
		c.JSON(http.StatusNoContent, nil)
	}
}

//Modify account information
//input params: name and role
func AccountRequestModifyHandler(c *gin.Context) {
	var r account.AccountRequest
	if err := c.Bind(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := c.Param("name")
	r.Name = name

	if err := account.AccountDB.Modify(r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		db.NotifyToSave()
		c.JSON(http.StatusNoContent, nil)
	}
}

//upgrade http to websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebTerminalHandler(c *gin.Context) {
	c.HTML(200, "terminal.html", nil)
}
