package main

import (
	"time"

	log "github.com/s00500/env_logger"
)

const localhost = true
const wifi = false

func main() {
	log.EnableLineNumbers()
	//tsp_test()
	getIp()
	log.Info("Got IP: ", ip)

	keyChan := make(chan string)
	go keyGet(keyChan)

	commandChan := make(chan string)
	poiChan := make(chan poiType)
	go initVisualServer(poiChan, commandChan)
	go initRobotServer(keyChan, poiChan, commandChan)

	//go UI.ServerInit()

	for {
		time.Sleep(2 * time.Second)
		//fmt.Println("goroutine: ", runtime.NumGoroutine())
	}
}
