package db

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/JinlongWukong/DevLab/account"
	"github.com/JinlongWukong/DevLab/config"
	"github.com/JinlongWukong/DevLab/manager"
	"github.com/JinlongWukong/DevLab/node"
	"github.com/JinlongWukong/DevLab/utils"
)

var requestChan = make(chan struct{}, 1)
var dbSyncPeriod = 5
var format = "json"

type DB struct {
}

var _ manager.Manager = DB{}

//initialize configuration
func init() {

	if config.DB.SyncPeriod > 0 {
		dbSyncPeriod = config.DB.SyncPeriod
	}
	if config.DB.Format != "" {
		format = config.DB.Format
	}

}

//Send notfication to DB chan to sync up
func NotifyToSave() {

	log.Println("Saving to file db")
	//non-blocking sends, keep only one request
	select {
	case requestChan <- struct{}{}:
	default:
	}
}

//Sync up map into db
func SaveToDB(ctx context.Context) {

	log.Println("Be ready to sync up with db")

	period := time.Duration(dbSyncPeriod) * time.Second
	t := time.NewTimer(period)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			select {
			case <-ctx.Done():
				return
			case <-requestChan:
				log.Println(time.Now())
				if format == "json" {
					err := utils.WriteJsonFile(".db/account.json", account.AccountDB.Map)
					if err != nil {
						log.Println(err)
					} else {
						log.Println("Saved to file db account.json")
					}
					err = utils.WriteJsonFile(".db/node.json", node.NodeDB.Map)
					if err != nil {
						log.Println(err)
					} else {
						log.Println("Saved to file db node.json")
					}
				} else {
					err := utils.GobStoreToFile(".db/account.db", account.AccountDB.Map)
					if err != nil {
						log.Println(err)
					} else {
						log.Println("Saved to file db account.db")
					}
					err = utils.GobStoreToFile(".db/node.db", node.NodeDB.Map)
					if err != nil {
						log.Println(err)
					} else {
						log.Println("Saved to file db node.db")
					}
				}

			}
			t.Reset(period)
		}
	}
}

//Load data from database
func LoadFromDB() {
	if format == "json" {
		accountData, err := utils.ReadJsonFile(".db/account.json")
		if err == nil {
			json.Unmarshal(accountData, &account.AccountDB.Map)
		} else if strings.Contains(err.Error(), "The system cannot find the file specified") ||
			strings.Contains(err.Error(), "no such file or directory") {
			log.Println("account.json db file not found, no content will be loaded")
		} else {
			log.Fatalf("account.json DB file load failed with error: %v", err)
		}

		nodeData, err := utils.ReadJsonFile(".db/node.json")
		if err == nil {
			json.Unmarshal(nodeData, &node.NodeDB.Map)
		} else if strings.Contains(err.Error(), "The system cannot find the file specified") ||
			strings.Contains(err.Error(), "no such file or directory") {
			log.Println("node.json db file not found, no content will be loaded")
		} else {
			log.Fatalf("node.json DB file load failed with error: %v", err)
		}
	} else {
		err := utils.GobLoadFromFile(".db/account.db", &account.AccountDB.Map)
		if err == nil {
			log.Println("account.db DB file loaded")
		} else if strings.Contains(err.Error(), "The system cannot find the file specified") {
			log.Println("account.db db file not found, no content will be loaded")
		} else {
			log.Fatalf("account.db DB file load failed with error: %v", err)
		}

		err = utils.GobLoadFromFile(".db/node.db", &node.NodeDB.Map)
		if err == nil {
			log.Println("node.db DB file loaded")
		} else if strings.Contains(err.Error(), "The system cannot find the file specified") {
			log.Println("node.db db file not found, no content will be loaded")
		} else {
			log.Fatalf("node.db DB file load failed with error: %v", err)
		}
	}
}

//DB controller
func (db DB) Control(ctx context.Context, wg *sync.WaitGroup) {

	log.Println("DB manager started")
	defer func() {
		log.Println("DB manager exited")
		wg.Done()
	}()

	//Load data from db into map
	LoadFromDB()

	account.AccountDB.InitializeAdmin()

	//for Loop to sync up db
	SaveToDB(ctx)

}
