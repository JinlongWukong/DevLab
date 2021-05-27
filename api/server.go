package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/JinlongWukong/CloudLab/config"
)

var apiHost = ""
var apiPort = 8088

func LoadConfig() {
	if config.ApiServer.Host != "" {
		apiHost = config.ApiServer.Host
	}
	if config.ApiServer.Port > 0 {
		apiPort = config.ApiServer.Port
	}
}

func Server() *http.Server {

	LoadConfig()

	srv := &http.Server{
		Addr:    apiHost + ":" + strconv.Itoa(apiPort),
		Handler: setupRouter(),
	}
	log.Printf("api server serve at %v", srv.Addr)
	go func() {
		// serve connections
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	return srv
}
