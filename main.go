package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JinlongWukong/CloudLab/api"
	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/deployer"
	"github.com/JinlongWukong/CloudLab/lifecycle"
	"github.com/JinlongWukong/CloudLab/notification"
	"github.com/JinlongWukong/CloudLab/scheduler"
	"github.com/JinlongWukong/CloudLab/workflow"
)

func main() {

	//Used for stop service gracefully
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	//Load config.ini
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("configuration file loadling failed %v, program exited", err)
	}

	deployer.LoadConfig()
	scheduler.LoadConfig()
	workflow.LoadConfig()

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
	srv := api.Server()

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
