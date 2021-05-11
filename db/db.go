package db

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/JinlongWukong/CloudLab/account"
	"github.com/JinlongWukong/CloudLab/node"
	"github.com/JinlongWukong/CloudLab/utils"
)

var database = "file"
var requestChan chan message

type message struct {
	collection string
	name       string
	action     string
}

//Init chan size
//use unbuffered channel for file db
//use buffered channel for mongo db
func init() {
	if database == "file" {
		requestChan = make(chan message)
	} else if database == "mongo" {
		requestChan = make(chan message, 1000)
	} else {
		log.Println("database not specified, use in-memory map")
	}
}

//Send notfication to DB chan to sync up
func NotifyToDB(collection string, name string, action string) {
	if database == "file" {
		log.Println("Saving to file db")
		//non-blocking sends, keep only one message received
		select {
		case requestChan <- message{collection, name, action}:
		default:
		}
	} else if database == "mongo" {
		log.Println("Saving to mongdo db")
		requestChan <- message{collection, name, action}
	} else {
		return
	}
}

//always sync up map into db
func SaveToDB() {
	log.Println("Be ready to sync up db")
	for request := range requestChan {
		if database == "file" {
			err := utils.WriteJsonFile("account.json", account.Account_db)
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Saved to file db account.json")
			}
			err = utils.WriteJsonFile("node.json", node.Node_db)
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Saved to file db node.json")
			}
		}
		if request.collection == "account" {
			//TODO
		} else if request.collection == "node" {
			//TODO
		}
	}
}

//Load data from database into map
func LoadFromDB() {
	if database == "file" {
		accountData, err := utils.ReadJsonFile("account.json")
		if err == nil {
			json.Unmarshal(accountData, &account.Account_db)
		} else if strings.Contains(err.Error(), "The system cannot find the file specified") {
			log.Println("account.json db file not found, no content will be load")
		} else {
			log.Fatalf("account.json DB file load failed with error: %v", err)
		}

		nodeData, err := utils.ReadJsonFile("node.json")
		if err == nil {
			json.Unmarshal(nodeData, &node.Node_db)
		} else if strings.Contains(err.Error(), "The system cannot find the file specified") {
			log.Println("node.json db file not found, no content will be load")
		} else {
			log.Fatalf("node.json DB file load failed with error: %v", err)
		}
	}
}

//DB manager
func Manager() {

	if database != "file" && database != "mongo" {
		log.Println("no database used, manager exited")
		return
	}

	//Load data from db into map
	LoadFromDB()

	//for Loop to sync up db
	go func() {
		var wg sync.WaitGroup
		for {
			wg.Add(1)
			go func() {
				defer wg.Done()
				SaveToDB()
			}()
			wg.Wait()
			log.Println("DB manager abnormal, try serve again")
			time.Sleep(time.Second)
		}
	}()

	return
}
