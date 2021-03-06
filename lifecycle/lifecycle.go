package lifecycle

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/JinlongWukong/DevLab/account"
	"github.com/JinlongWukong/DevLab/config"
	"github.com/JinlongWukong/DevLab/db"
	"github.com/JinlongWukong/DevLab/manager"
	"github.com/JinlongWukong/DevLab/vm"
	"github.com/JinlongWukong/DevLab/workflow"
)

var checkInterval = "1h"
var enabled = false

//31536000000000000 = 1 years, if >= 1 years means forever
var forever = time.Duration(31536000000000000)

type LifeCycle struct {
}

var _ manager.Manager = LifeCycle{}

//initialize configuration
func init() {

	if config.LifeCycle.CheckInterval != "" {
		checkInterval = config.LifeCycle.CheckInterval
	}
	if config.LifeCycle.Enable == "true" {
		enabled = true
	}
	if config.LifeCycle.Forever > 0 {
		forever = time.Duration(config.LifeCycle.Forever)
	}
}

func (l LifeCycle) Control(ctx context.Context, wg *sync.WaitGroup) {

	log.Println("LifeCycle manager started")
	defer func() {
		log.Println("Lifecycle manager exited")
		wg.Done()
	}()

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
			case <-ctx.Done():
				return
			case <-t.C:
				for ac := range account.AccountDB.Iter() {
					vmSlice := []*vm.VirtualMachine{}
					for item := range ac.Value.Iter() {
						vmSlice = append(vmSlice, item)
					}
					for _, vm := range vmSlice {
						//whether forever
						if vm.GetLifeTime() >= forever {
							continue
						}
						vmLifeTime := vm.ChangeLifeTime(-period)
						log.Printf("Accout %v vm %v lifetime is %v", ac.Value.Name, vm.Name, vmLifeTime)
						if vmLifeTime <= 0 {
							log.Printf("%v Lifetime is over, begin to delete vm", vm.Name)
							if err := workflow.ActionVM(ac.Value, vm, "delete"); err != nil {
								log.Println(err)
							}
						} else if vmLifeTime < 6*time.Hour {
							ac.Value.SendNotification(fmt.Sprintf("Warning, Your VM %v still have %v life left", vm.Name, vmLifeTime))
						}
					}
					db.NotifyToSave()
				}
			}
		}
	}
}
