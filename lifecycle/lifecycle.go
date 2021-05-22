package lifecycle

import (
	"fmt"
	"log"
	"time"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/db"
	"github.com/JinlongWukong/CloudLab/workflow"
)

var checkInterval = "1h"
var enabled = false

func initialize() {

	if config.LifeCycle.CheckInterval != "" {
		checkInterval = config.LifeCycle.CheckInterval
	}
	if config.LifeCycle.Enable == "true" {
		enabled = true
	}
}

func Manager() {

	initialize()

	go func() {
		if enabled == true {
			log.Println("Lifecycle is enabled, begin to work")
			period, err := time.ParseDuration(checkInterval)
			if err != nil {
				log.Println(err)
				return
			}
			t := time.NewTicker(period)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					for ac := range account.AccountDB.Iter() {
						for _, vm := range ac.Value.VM {
							vm.Lifetime = vm.Lifetime - period
							log.Printf("Accout %v vm %v lifetime is %v", ac.Value.Name, vm.Name, vm.Lifetime)
							if vm.Lifetime <= 0 {
								log.Printf("%v Lifetime is over, begin to delete vm", vm.Name)
								if err := workflow.ActionVM(ac.Value, vm, "delete"); err != nil {
									log.Println(err)
								}
							} else if vm.Lifetime < 6*time.Hour {
								ac.Value.SendNotification(fmt.Sprintf("Warning, Your VM %v still have %v life left", vm.Name, vm.Lifetime))
							}
						}
						db.NotifyToDB("account", ac.Value.Name, "update")
					}
				}
			}
		}
	}()

}