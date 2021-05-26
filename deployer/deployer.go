package deployer

import (
	"log"
	"os"
	"regexp"

	"github.com/JinlongWukong/CloudLab/config"
)

var baseUrl string

func ReloadConfig() {
	envUrl := os.Getenv("DEPLOYER_PROTOCOL") + "://" + os.Getenv("DEPLOYER_HOST") + ":" + os.Getenv("DEPLOYER_PORT")
	match, _ := regexp.MatchString("^http[s]?://[[:ascii:]]*:\\d*$", envUrl)
	if match {
		log.Printf("Using deployer url from environment %v", envUrl)
		baseUrl = envUrl
	} else {
		baseUrl = config.Deployer.Protocol + "://" + config.Deployer.EndPoint
		log.Printf("Using deployer url from config.ini %v", baseUrl)
	}
}

func GetDeployerBaseUrl() string {
	return baseUrl
}
