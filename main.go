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
	//vm request first page
	r.GET("/vm-home", api.VmRequestIndexHandler)
	//Get vm http://127.0.0.1:8088/vm?account=1234
	r.GET("/vm", api.VmRequestGetVmHandler)
	//Create vm
	r.POST("/vm", api.VmRequestCreateVmHandler)
	//Generate vm action(start/stop/reboot/delete)
	r.POST("/vm/action", api.VmRequestVmActionHandler)
	//Port expose
	r.POST("/vm/expose-port", api.VmRequestVmPortExposeHandler)

	//node related api
	r.GET("/node", api.NodeRequestGetNodeHandler)
	r.POST("/node", api.NodeRequestActionNodeHandler)

	//TODO api
	r.GET("/k8s", api.ToDoHandler)
	r.GET("/container", api.ToDoHandler)

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
