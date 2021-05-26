package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/JinlongWukong/CloudLab/api"
	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/deployer"
	"github.com/JinlongWukong/CloudLab/lifecycle"
	"github.com/JinlongWukong/CloudLab/notification"
	"github.com/JinlongWukong/CloudLab/scheduler"
	"github.com/JinlongWukong/CloudLab/workflow"
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

	r.GET("/", api.IndexHandler)
	r.GET("/admin", api.AdminHandler)

	//vm related api
	//vm request first page
	r.GET("/vm-request", api.VmRequestIndexHandler)
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

	return r
}

func main() {

	//Used for stop service
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	//Load config.ini
	config.LoadConfig()

	deployer.ReloadConfig()
	scheduler.ReloadConfig()
	workflow.ReloadConfig()

	//Start db control loop
	wg.Add(1)
	go db.Manager(ctx, &wg)

	//Start notification loop
	wg.Add(1)
	go notification.Manager(ctx, &wg)

	//Start lifecycle loop
	wg.Add(1)
	go lifecycle.Manager(ctx, &wg)

	//Setup web server
	srv := &http.Server{
		Addr:    ":8088",
		Handler: setupRouter(),
	}
	go func() {
		// serve connections
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	//Wait signal to reload/stop
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM)
	for s := range sigs {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			ctxServer, cancelServer := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelServer()
			if err := srv.Shutdown(ctxServer); err != nil {
				log.Printf("Server Shutdown: %v", err)
			}
			cancel()
			ctxManger, cancelManager := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelManager()
			go func() {
				<-ctxManger.Done()
				log.Println("Manger exit timeout")
				os.Exit(1)
			}()
			wg.Wait()
			log.Println("Program exit normally")
			os.Exit(0)
		}
	}
}
