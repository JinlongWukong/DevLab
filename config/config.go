package config

import (
	"log"

	"gopkg.in/ini.v1"
)

type DatabaseConfig struct {
	//database sync up period
	SyncPeriod int
	//database format(json, gob)
	Format string
}

type NotificationConfig struct {
	//notification kind(webex,...)
	Kind string
	//notification queue size
	QueueSize int
}

type ScheduleConfig struct {
	AllocationRatio   int
	ScheduleAlgorithm string
}

type WorkflowConfig struct {
	VmStatusRetry    int
	VmStatusInterval int
}

type LifeCycleConfig struct {
	CheckInterval string
	Enable        string
	Forever       int64
}

type DeployerConfig struct {
	Protocol string
	EndPoint string
}

type ApiServerConfig struct {
	Host string
	Port int
}

type SupervisorConfig struct {
	//Enable/disable supervisor
	Enable string
	//Node check interval -> 10s,1m
	NodeCheckInterval string
	//Maxinum CPU load(15mins) usage, 0.8 -> core * 80%
	NodeLimitCPU float64
	//Mininum memory available left, unit(M)
	NodeMinimumMem int
	//maxinum Disk in Use(0,100), 80 -> 80%
	NodeLimitDisk int
}

type NodeConfig struct {
	//Node subnet range
	SubnetRange string
}

type NetworkConfig struct {
	CheckInterval string
	NetworkType   string
}

var DB DatabaseConfig
var Workflow WorkflowConfig
var Schedule ScheduleConfig
var Notification NotificationConfig
var LifeCycle LifeCycleConfig
var Deployer DeployerConfig
var ApiServer ApiServerConfig
var Supervisor SupervisorConfig
var Node NodeConfig
var Network NetworkConfig

func init() {

	//Load config.ini
	if err := LoadConfig(); err != nil {
		log.Fatalf("configuration file loadling failed %v, program exited", err)
	}

}

func LoadConfig() error {

	log.Println("Start loading config.ini")
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Printf("Fail to read file: %v", err)
		return err
	}

	err = cfg.Section("Database").MapTo(&DB)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Database", err)
		return err
	}

	err = cfg.Section("Workflow").MapTo(&Workflow)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Workflow", err)
		return err
	}
	err = cfg.Section("Schedule").MapTo(&Schedule)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Schedule", err)
		return err
	}

	err = cfg.Section("Notification").MapTo(&Notification)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Notification", err)
		return err
	}

	err = cfg.Section("Lifecycle").MapTo(&LifeCycle)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Lifecycle", err)
		return err
	}

	err = cfg.Section("Deployer").MapTo(&Deployer)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Deployer", err)
		return err
	}

	err = cfg.Section("ApiServer").MapTo(&ApiServer)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "ApiServer", err)
		return err
	}

	err = cfg.Section("Supervisor").MapTo(&Supervisor)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Supervisor", err)
		return err
	}

	err = cfg.Section("Node").MapTo(&Node)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Node", err)
		return err
	}

	err = cfg.Section("Network").MapTo(&Network)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Network", err)
		return err
	}

	log.Println("All configuration loading done")
	return nil

}
