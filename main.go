package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JinlongWukong/DevLab/api"
	"github.com/JinlongWukong/DevLab/db"
	"github.com/JinlongWukong/DevLab/lifecycle"
	"github.com/JinlongWukong/DevLab/manager"
	"github.com/JinlongWukong/DevLab/network"
	"github.com/JinlongWukong/DevLab/notification"
	"github.com/JinlongWukong/DevLab/supervisor"
)

func main() {

	//Used for stop service gracefully
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	//Setup all managers
	managers := make([]manager.Manager, 0)
	managers = append(managers,
		db.DB{},
		notification.Notifier{},
		lifecycle.LifeCycle{},
		supervisor.Supervisor{},
		network.NetworkController{},
	)
	for _, m := range managers {
		wg.Add(1)
		go m.Control(ctx, &wg)
	}

	//Setup web server
	srv := api.Server()

	//Wait signal to stop service gracefully
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
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
