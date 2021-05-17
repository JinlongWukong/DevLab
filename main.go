package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/JinlongWukong/CloudLab/api"
	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/notification"
)

func main() {

	//Load config.ini
	config.Manager()

	//Start db control loop
	db.Manager()

	//Start notification loop
	notification.Manager()

	r := gin.Default()
	r.Use(cors.Default())

	r.LoadHTMLGlob("views/*")
	r.GET("/", api.IndexHandler)
	r.GET("/admin", api.AdminHandler)

	//vm related api
	r.GET("/vm-request", api.VmRequestIndexHandler)
	r.GET("/vm-request/vm", api.VmRequestGetVmHandler)

	r.POST("/vm-request", api.VmRequestCreateVmHandler)
	r.POST("/vm-request/vm", api.VmRequestVmActionHandler)

	//node related api
	r.GET("/node-request", api.NodeRequestGetNodeHandler)
	r.POST("/node-request", api.NodeRequestActionNodeHandler)

	//TODO api
	r.GET("/k8s-request", api.ToDoHandler)
	r.GET("/container-request", api.ToDoHandler)

	r.Run(":8088")
}
