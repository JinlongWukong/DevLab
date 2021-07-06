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
	r.Static("/scripts", "views/scripts")
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
	r.GET("/vm", AuthorizeToken(), VmRequestGetVmHandler)
	//Create vm
	r.POST("/vm", AuthorizeToken(), VmRequestCreateVmHandler)
	//Generate vm action(start/stop/reboot/delete)
	r.POST("/vm/action", AuthorizeToken(), VmRequestVmActionHandler)
	//Port expose
	r.POST("/vm/expose-port", AuthorizeToken(), VmRequestVmPortExposeHandler)

	//node related api
	r.GET("/node", AuthorizeToken(), NodeRequestGetNodeHandler)
	r.POST("/node", AuthorizeToken(), AdminRoleOnlyAllowed(), NodeRequestActionNodeHandler)

	//account related api
	r.POST("/account", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestPostHandler)
	r.GET("/account", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestGetAllHandler)
	r.GET("/account/:name", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestGetByNameHandler)
	r.PATCH("/account/:name", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestModifyHandler)
	r.DELETE("/account/:name", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestDelByNameHandler)

	//workflow related api
	r.GET("/task", WorkflowTaskHandler)

	//k8s related api
	r.GET("/k8s-request", K8sRequestIndexHandler)
	r.POST("/k8s", AuthorizeToken(), K8sRequestCreateHandler)
	r.DELETE("/k8s", AuthorizeToken(), K8sRequestDeleteHandler)
	r.GET("/k8s", AuthorizeToken(), K8sRequestGetHandler)

	//SaaS api
	r.GET("/saas-request", SoftwareIndexHandler)
	r.GET("/saas", AuthorizeToken(), SoftwareRequestGetHandler)
	r.POST("/saas", AuthorizeToken(), SoftwareRequestCreateHandler)
	r.POST("/saas/action", AuthorizeToken(), SoftwareRequestActionHandler)

	//auth api
	r.POST("/one-time-password", oneTimePassGenHandler)
	r.POST("/login", accountLogin)

	return r
}
