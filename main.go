package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/JinlongWukong/CloudLab/api"
	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/lifecycle"
	"github.com/JinlongWukong/CloudLab/notification"
)

func routeRegister(r *gin.Engine) {

	r.LoadHTMLGlob("views/*")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", api.IndexHandler)
	r.GET("/admin", api.AdminHandler)

	//vm related api
	r.GET("/vm-request", api.VmRequestIndexHandler)
	r.GET("/vm-request/vm", api.VmRequestGetVmHandler)

	r.POST("/vm-request", api.VmRequestCreateVmHandler)
	r.POST("/vm-request/vm", api.VmRequestVmActionHandler)
	r.POST("/vm-request/vm/port-expose", api.VmRequestVmPortExposeHandler)

	//node related api
	r.GET("/node-request", api.NodeRequestGetNodeHandler)
	r.POST("/node-request", api.NodeRequestActionNodeHandler)

	//TODO api
	r.GET("/k8s-request", api.ToDoHandler)
	r.GET("/container-request", api.ToDoHandler)

}

func main() {

	//Load config.ini
	config.Manager()

	//Start db control loop
	db.Manager()

	//Start notification loop
	notification.Manager()

	//Start lifecycle loop
	lifecycle.Manager()

	//Start web server
	r := gin.Default()
	r.Use(cors.Default())
	routeRegister(r)
	r.Run(":8088")
}
