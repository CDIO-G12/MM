package main

import (
	f "MM/frame"
	u "MM/utils"
	"time"

	log "github.com/s00500/env_logger"
)

func main() {
	// enable line numbers in log
	log.EnableLineNumbers()
	//tsp_test()
	u.GetIp()
	log.Info("Got IP: ", u.IP)

	// this creates a channel and a routine for the key getter. Not used at the moment
	keyChan := make(chan string)
	go u.KeyGet(keyChan)

	// channels between visuals and robot server
	pixelDist := &u.PixelDistType{}
	commandChan := make(chan string)
	poiChan := make(chan u.PoiType)
	frame := f.NewFrame(poiChan, pixelDist)
	go initVisualServer(poiChan, commandChan, pixelDist)
	go initRobotServer(frame, keyChan, poiChan, commandChan, pixelDist)

	// loop main, can enable prints of routines to watch out for too many go routines
	for {
		time.Sleep(2 * time.Second)
		//fmt.Println("goroutine: ", runtime.NumGoroutine())
	}
}
