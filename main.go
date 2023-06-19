package main

import (
	ct "MM/conn_testers"
	f "MM/frame"
	u "MM/utils"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	log "github.com/s00500/env_logger"
)

var Timer u.TimerType

// CLICOLOR_FORCE=1 go run .
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// enable line numbers in log
	log.EnableLineNumbers()

	u.GetIp()
	log.Info("Got IP: ", u.IP)

	Timer = u.NewTimer()

	// this creates a channel and a routine for the key getter. Not used at the moment
	keyChan := make(chan string)
	go u.KeyGet(keyChan)

	// channels between visuals and robot server
	commandChan := make(chan string, 15)
	poiChan := make(chan u.PoiType, 15)
	framePoiChan := make(chan u.PoiType, 50)
	frame := f.NewFrame(framePoiChan)
	go initVisualServer(frame, poiChan, framePoiChan, commandChan)
	go initRobotServer(frame, keyChan, poiChan, commandChan)

	// loop main
	buf := ""
	for {
		fmt.Scanln(&buf)
		switch buf {
		case "time":
			log.Infoln("Time: ", Timer.Now())
			continue
		case "left":
			log.Infoln("Time left: ", Timer.Left())
		case "r":
			log.Infoln("Go routines: ", runtime.NumGoroutine())
		case "cal":
			poiChan <- u.PoiType{Category: u.Calibrate}
		case "exit":
			pprof.StopCPUProfile()
			log.Fatal("Exiting")
		default:
			log.Infof("Got '%s' from terminal", buf)
			commandChan <- buf
		}
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
