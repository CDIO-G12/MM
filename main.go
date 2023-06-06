package main

import (
	ct "MM/conn_testers"
	f "MM/frame"
	u "MM/utils"
	"fmt"
	"strings"
	"time"

	log "github.com/s00500/env_logger"
)

//CLICOLOR_FORCE=1 go run .

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

	// to test with no robot
	//robotFaker()

	// loop main, can enable prints of routines to watch out for too many go routines
	for {
		time.Sleep(2 * time.Second)
		//log.Info("goroutine: ", runtime.NumGoroutine())
	}
}

func robotFaker() {
	robotIn := make(chan string)
	robotOut := make(chan string)
	time.Sleep(5 * time.Second)
	go ct.RobotClient(robotIn, robotOut)
	robotIn <- "rd"
	for {
		out := <-robotOut
		fmt.Printf("Robot got: %c, %d\n", out[0], int(out[1]))
		time.Sleep(1 * time.Second)
		switch {

		case strings.Contains(out, "D"):
			robotIn <- "dc"
			continue

		case strings.Contains(out, "S"), strings.Contains(out, "T"):
			robotIn <- "gb"
			//visChan <- currentPosToString(currentPos)
			continue
		}

		//visChan <- currentPosToString(currentPos)
		time.Sleep(10 * time.Millisecond)

		robotIn <- "fm"

	}
}
