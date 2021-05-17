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

var DB DatabaseConfig
var Workflow WorkflowConfig
var Schedule ScheduleConfig
var Notification NotificationConfig

func Manager() {

	log.Println("Start loading config.ini")
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Printf("Fail to read file: %v", err)
	}

	err = cfg.Section("Database").MapTo(&DB)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Database", err)
	}

	err = cfg.Section("Workflow").MapTo(&Workflow)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Workflow", err)
	}

	err = cfg.Section("Schedule").MapTo(&Schedule)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Schedule", err)
	}

	err = cfg.Section("Notification").MapTo(&Notification)
	if err != nil {
		log.Printf("Fail to parse section %v: %v", "Notification", err)
	}

	log.Println("All configuration loading done")
}
