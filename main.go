package main

import (
	"time"

	log "github.com/s00500/env_logger"
)

const localhost = false
const wifi = false

func main() {
	// enable line numbers in log
	log.EnableLineNumbers()
	//tsp_test()
	getIp()
	log.Info("Got IP: ", ip)

	// this creates a channel and a routine for the key getter. Not used at the moment
	keyChan := make(chan string)
	go keyGet(keyChan)

	// channels between visuals and robot server
	commandChan := make(chan string)
	poiChan := make(chan poiType)
	go initVisualServer(poiChan, commandChan)
	go initRobotServer(keyChan, poiChan, commandChan)

	// loop main, can enable prints of routines to watch out for too many go routines
	for {
		time.Sleep(2 * time.Second)
		//fmt.Println("goroutine: ", runtime.NumGoroutine())
	}
}
