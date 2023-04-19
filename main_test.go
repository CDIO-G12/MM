package main

import (
	ct "MM/conn_testers"
	f "MM/frame"
	u "MM/utils"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	log "github.com/s00500/env_logger"
)

func Test_main(t *testing.T) {
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
	commandChan := make(chan string)
	poiChan := make(chan u.PoiType)
	frame := f.NewFrame(poiChan)
	go initVisualServer(poiChan, commandChan)
	go initRobotServer(frame, keyChan, poiChan, commandChan)

	visChan := make(chan string)
	robotIn := make(chan string)
	robotOut := make(chan string)
	currentPos := u.PointType{X: 100, Y: 100, Angle: 0}

	go ct.VisualsClient(visChan)
	go ct.RobotClient(robotIn, robotOut)

	go func() {
		for {
			out := <-robotOut
			fmt.Printf("Robot got: %c, %d\n", out[0], int(out[1]))
			time.Sleep(1 * time.Second)
			switch {
			case strings.Contains(out, "L"):
				currentPos.Angle = u.DegreeAdd(currentPos.Angle, -int(out[1]))

			case strings.Contains(out, "R"):
				currentPos.Angle = u.DegreeAdd(currentPos.Angle, int(out[1]))

			case strings.Contains(out, "f"):
				currentPos = currentPos.CalcNextPos(int(out[1]))

			case strings.Contains(out, "F"):
				currentPos = currentPos.CalcNextPos(int(out[1]) * 255)

			case strings.Contains(out, "B"):
				currentPos = currentPos.CalcNextPos(-int(out[1]))
			}

			visChan <- currentPosToString(currentPos)
			time.Sleep(10 * time.Millisecond)

			robotIn <- "fm"

		}
	}()

	time.Sleep(100 * time.Millisecond)
	initializeVisuals(visChan, currentPos)
	time.Sleep(100 * time.Millisecond)
	robotIn <- "rd"
	time.Sleep(20 * time.Second)
	fmt.Println("goroutine: ", runtime.NumGoroutine())

}

func initializeVisuals(visChan chan<- string, currentPos u.PointType) {
	visChan <- "c/0/25/25"
	visChan <- "c/1/250/25"
	visChan <- "c/2/250/125"
	visChan <- "c/3/25/125"
	for i := 0; i < 2; i++ {
		visChan <- "b/r/r"
		visChan <- "b/50/50"
		visChan <- "b/200/40"
		visChan <- "b/250/120"
		visChan <- "b/d/d"
	}
	visChan <- currentPosToString(currentPos)

}

func currentPosToString(currentPos u.PointType) string {
	return fmt.Sprintf("r/%d/%d/%d", currentPos.X, currentPos.Y, u.DegreeAdd(currentPos.Angle, 90))
}
