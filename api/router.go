package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {

	r := gin.Default()
	r.Use(cors.Default())

	r.LoadHTMLGlob("views/*")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", IndexHandler)
	r.GET("/admin", AdminHandler)

	//vm related api
	//vm request first page
	r.GET("/vm-request", VmRequestIndexHandler)
	//Get vm http://127.0.0.1:8088/vm?account=1234
	r.GET("/vm", VmRequestGetVmHandler)
	//Create vm
	r.POST("/vm", VmRequestCreateVmHandler)
	//Generate vm action(start/stop/reboot/delete)
	r.POST("/vm/action", VmRequestVmActionHandler)
	//Port expose
	r.POST("/vm/expose-port", VmRequestVmPortExposeHandler)

	//node related api
	r.GET("/node", NodeRequestGetNodeHandler)
	r.POST("/node", NodeRequestActionNodeHandler)

	//TODO api
	r.GET("/k8s", ToDoHandler)
	r.GET("/container", ToDoHandler)

	return r
}
