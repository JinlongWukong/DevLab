package config

import (
	"log"

	"gopkg.in/ini.v1"
)

type DatabaseConfig struct {
	Database     string
	DBSyncPeriod int
}

type NotificationConfig struct {
	Kind      string
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
}

type DeployerConfig struct {
	Protocol string
	EndPoint string
}

type ApiServerConfig struct {
	Host string
	Port int
}

var DB DatabaseConfig
var Workflow WorkflowConfig
var Schedule ScheduleConfig
var Notification NotificationConfig
var LifeCycle LifeCycleConfig
var Deployer DeployerConfig
var ApiServer ApiServerConfig

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

	log.Println("All configuration loading done")
	return nil

}
