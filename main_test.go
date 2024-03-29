package main

/*
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
	t.Skip()

	// enable line numbers in log
	log.EnableLineNumbers()
	//tsp_test()
	u.GetIp()
	log.Info("Got IP: ", u.IP)



	// this creates a channel and a routine for the key getter. Not used at the moment
	keyChan := make(chan string, 5)
	//go u.KeyGet(keyChan)

	// channels between visuals and robot server
	commandChan := make(chan string)
	poiChan := make(chan u.PoiType)
	framePoiChan := make(chan u.PoiType)
	frame := f.NewFrame(framePoiChan)

	go initVisualServer(frame, poiChan, framePoiChan, commandChan)
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
				d := float64(int(out[1])) / u.GetPixelDist()
				currentPos = currentPos.CalcNextPos(int(d))

			case strings.Contains(out, "F"):
				d := float64(int(out[1])) / u.GetPixelDist()
				fmt.Println(d)
				currentPos = currentPos.CalcNextPos(int(d) * 10)

			case strings.Contains(out, "B"):
				d := float64(int(out[1])) / u.GetPixelDist()
				currentPos = currentPos.CalcNextPos(-int(d))

			case strings.Contains(out, "D"):
				robotIn <- "dc"
				continue

			case strings.Contains(out, "S"), strings.Contains(out, "T"):
				robotIn <- "gb"
				visChan <- currentPosToString(currentPos)
				continue
			}
			u.CurrentPos.Set(currentPos)

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
		visChan <- "b/51/50"
		visChan <- "b/200/40"
		visChan <- "b/250/120"
		visChan <- "b/d/d"
	}
	visChan <- currentPosToString(currentPos)

}

func currentPosToString(currentPos u.PointType) string {
	return fmt.Sprintf("r/%d/%d/%d", currentPos.X, currentPos.Y, u.DegreeAdd(currentPos.Angle, 90))
}
*/
