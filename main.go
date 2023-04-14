package main

import (
	"sync"
	"time"

	log "github.com/s00500/env_logger"
)

var pixelDistMU = sync.RWMutex{}
var pixelDist = 10.0

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
	frame := newFrame(poiChan)
	go initVisualServer(poiChan, commandChan)
	go initRobotServer(frame, keyChan, poiChan, commandChan)

	// loop main, can enable prints of routines to watch out for too many go routines
	for {
		time.Sleep(2 * time.Second)
		//fmt.Println("goroutine: ", runtime.NumGoroutine())
	}
}
