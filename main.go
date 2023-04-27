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

	/*f.ManualTest()
	return
	*/

	// this creates a channel and a routine for the key getter. Not used at the moment
	keyChan := make(chan string)
	go u.KeyGet(keyChan)

	// channels between visuals and robot server
	commandChan := make(chan string, 5)
	poiChan := make(chan u.PoiType, 5)
	framePoiChan := make(chan u.PoiType, 5)
	frame := f.NewFrame(framePoiChan)
	go initVisualServer(frame, poiChan, framePoiChan, commandChan)
	go initRobotServer(frame, keyChan, poiChan, commandChan)

	robotIn := make(chan string)
	//robotOut := make(chan string)

	time.Sleep(5 * time.Second)

	//go ct.RobotClient(robotIn, robotOut)
	robotIn <- "rd"

	// loop main, can enable prints of routines to watch out for too many go routines
	for {
		time.Sleep(2 * time.Second)
		//fmt.Println("goroutine: ", runtime.NumGoroutine())
	}
}
