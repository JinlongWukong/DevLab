package deployer

import (
	"log"
	"os"
	"regexp"

	"github.com/JinlongWukong/CloudLab/config"
)

var baseUrl string

func ReloadConfig() {
	//Load deployer info from ENV first, if not ok, then load from config.ini
	envUrl := os.Getenv("DEPLOYER_PROTOCOL") + "://" + os.Getenv("DEPLOYER_HOST") + ":" + os.Getenv("DEPLOYER_PORT")
	match, _ := regexp.MatchString("^http[s]?://[[:ascii:]]*:\\d*$", envUrl)
	if match {
		log.Printf("Using deployer url from environment %v", envUrl)
		baseUrl = envUrl
		return
	}

	configUrl := config.Deployer.Protocol + "://" + config.Deployer.EndPoint
	match, _ = regexp.MatchString("^http[s]?://[[:ascii:]]*:\\d*$", configUrl)
	if match {
		baseUrl = config.Deployer.Protocol + "://" + config.Deployer.EndPoint
		log.Printf("Using deployer url from config.ini %v", baseUrl)
		return
	}
	log.Println("deployer url info load failed")
}

func GetDeployerBaseUrl() string {
	return baseUrl
}
