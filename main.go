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

// VM reqeust POST handler
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
