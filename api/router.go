package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {

	r := gin.Default()
	r.Use(cors.Default())

	r.Static("/css", "views/css")
	r.Static("/img", "views/image")
	r.LoadHTMLGlob("views/*.html")
	//readness/liveness check point
	r.GET("/ping", func(c *gin.Context) {
		c.Writer.WriteString("pong")
	})
	//metrics
	r.GET("/internal/metrics", metricsHandler)

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

	//workflow related api
	r.GET("/task", WorkflowTaskHandler)

	//k8s related api
	r.GET("/k8s-request", K8sRequestIndexHandler)
	r.POST("/k8s", K8sRequestCreateHandler)
	r.DELETE("/k8s", K8sRequestDeleteHandler)
	r.GET("/k8s", K8sRequestGetHandler)

	//SaaS api
	r.GET("/saas-request", SoftwareIndexHandler)
	r.GET("/saas", SoftwareRequestGetHandler)
	r.POST("/saas", SoftwareRequestCreateHandler)
	r.POST("/saas/action", SoftwareRequestActionHandler)

	return r
}
