package main

import (
	"log"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/vm"
	"github.com/JinlongWukong/CloudLab/workflow"
	"github.com/gin-gonic/gin"
)

var account_db = make(map[string]*account.Account)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("views/*")
	r.GET("/", indexHandler)

	r.GET("/vm-request", vmRequestGetHandler)
	r.POST("/vm-request", vmRequestPostHandler)

	r.Run(":8088")
}

// First Page
func indexHandler(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

// VM reqeust GET handler
func vmRequestGetHandler(c *gin.Context) {
	c.HTML(200, "vmRequest.html", nil)
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
