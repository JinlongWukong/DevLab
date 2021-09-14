package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {

	r := gin.Default()
	r.Use(cors.Default())

	//static files
	r.Static("/css", "views/css")
	r.Static("/fonts", "views/fonts")
	r.Static("/img", "views/image")
	r.Static("/scripts", "views/scripts")
	r.Delims("{{{", "}}}")
	r.LoadHTMLGlob("views/*.html")

	//readness/liveness check point
	r.GET("/ping", func(c *gin.Context) {
		c.Writer.WriteString("pong")
	})

	//metrics
	r.GET("/internal/metrics", metricsHandler)

	//workflow related api
	r.GET("/task", WorkflowTaskHandler)

	//home page
	r.GET("/", IndexHandler)

	//admin page(node, account management)
	r.GET("/admin", AdminHandler)
	r.GET("/admin-node", AdminNodeHandler)
	r.GET("/admin-account", AdminAccountHandler)

	//node related api
	r.GET("/node", AuthorizeToken(), NodeRequestGetAllHandler)
	r.GET("/node/:name", AuthorizeToken(), NodeRequestGetByNameHandler)
	r.POST("/node", AuthorizeToken(), AdminRoleOnlyAllowed(), NodeRequestCreateHandler)
	r.POST("/node/:name/:action", AuthorizeToken(), AdminRoleOnlyAllowed(), NodeRequestActionHandler)

	//account related api
	r.POST("/account", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestCreateHandler)
	r.GET("/account", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestGetAllHandler)
	r.GET("/account/:name", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestGetByNameHandler)
	r.PATCH("/account/:name", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestModifyHandler)
	r.DELETE("/account/:name", AuthorizeToken(), AdminRoleOnlyAllowed(), AccountRequestDelByNameHandler)

	//vm related api
	r.GET("/vm-request", VmRequestIndexHandler)
	r.GET("/vm", AuthorizeToken(), VmRequestGetAllHandler)
	r.GET("/vm/:name", AuthorizeToken(), VmRequestGetByNameHandler)
	r.POST("/vm", AuthorizeToken(), VmRequestCreateHandler)
	r.POST("/vm/:name/:action", AuthorizeToken(), VmRequestActionHandler)
	r.POST("/vm/:name/port/expose", AuthorizeToken(), VmRequestPortExposeHandler)
	r.GET("/vm/:name/ws", VmRequestWebConsole)
	r.GET("/vm/:name/web-terminal", WebTerminalHandler)

	//k8s related api
	r.GET("/k8s-request", K8sRequestIndexHandler)
	r.POST("/k8s", AuthorizeToken(), K8sRequestCreateHandler)
	r.DELETE("/k8s/:name", AuthorizeToken(), K8sRequestDeleteHandler)
	r.GET("/k8s", AuthorizeToken(), K8sRequestGetAllHandler)
	r.GET("/k8s/:name", AuthorizeToken(), K8sRequestGetByNameHandler)

	//SaaS related api
	r.GET("/saas-request", SoftwareIndexHandler)
	r.GET("/saas", AuthorizeToken(), SoftwareRequestGetAllHandler)
	r.GET("/saas/:name", AuthorizeToken(), SoftwareRequestGetByNameHandler)
	r.POST("/saas", AuthorizeToken(), SoftwareRequestCreateHandler)
	r.POST("/saas/:name/:action", AuthorizeToken(), SoftwareRequestActionHandler)
	r.GET("/container/:name/ws", ContainerRequestWebConsole)
	r.GET("/container/:name/web-terminal", WebTerminalHandler)

	//auth api
	r.POST("/one-time-password", oneTimePassGenHandler)
	r.POST("/login", accountLoginHandler)

	//web ssh terminal
	//r.GET("/ws/:host/:port/:user/:password", WebConsole)
	//r.GET("/web-terminal/:host/:port/:user/:password", WebTerminalHandler)
	return r
}
