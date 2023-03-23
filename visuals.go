package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/s00500/env_logger"
)

const visualPort = 8888

// initVisualServer hold the visual server and handles stuff
func initVisualServer(poiChan chan<- poiType, commandChan chan string) {
	log.Info("Visual server started")
	//go imageReciever()
	balls := []pointType{}
	sortedBalls := []pointType{}
	currentPos := pointType{x: 200, y: 200, angle: 180}
	goalPos := pointType{x: 200, y: 400}
	active := false

	// this routine handles commands comming from the robot
	go func() {
		for {
			// if it is not active, it will ignore commands from the robot
			if !active {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// recieve command from robot, and handle
			cmd := <-commandChan
			switch cmd {
			case "first": // Send first ball
				log.Infoln("First ball send")
				if len(sortedBalls) > 0 {
					poiChan <- poiType{point: sortedBalls[0], category: ball}
				} else {
					// do something if it has not found balls yet
				}

			case "next": // Send next ball
				if len(sortedBalls) == 0 {
					poiChan <- poiType{point: goalPos, category: goal}
					log.Infoln("Goal send")
				} else {
					// channel sends blocks, so we open a new routine to not stop the current
					go func() {
						commandChan <- "compute"
					}()
				}

			case "pos": // Send current position
				poiChan <- poiType{point: currentPos, category: robot}

			case "compute": // compute ball positions
				var err error
				tempBalls := make([]pointType, len(balls))
				copy(tempBalls, balls)
				sortedBalls, err = currentPos.sortBalls(tempBalls)
				if err != nil {
					return
				}
				log.Info("Computed balls: ", sortedBalls)
				poiChan <- poiType{point: sortedBalls[0], category: ball}

			case "goal":
				poiChan <- poiType{point: goalPos, category: goal}
			}
		}
	}()

	// create server
	addr := fmt.Sprintf("%s:%d", ip, visualPort)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	log.Println("Visuals server is running on:", addr)
	for {
		// Accept incoming server request
		conn, err := server.Accept()
		if err != nil {
			log.Println("Failed to accept conn.", err)
			continue
		}
		log.Infoln("Connected to visuals at:", conn.RemoteAddr().String())
		buffer := make([]byte, 128)
		ballBuffer := []pointType{}
		active = true

		for {
			// Read blocks, so we wait for incoming command
			rLen, err := conn.Read(buffer)
			// if it fails, we break the loop
			if log.Should(err) {
				// Send an emergency in a new routine so it wont block
				go func() {
					poiChan <- poiType{category: emergency} // Stop the bot if connection is lost
				}()
				conn.Close()
				break
			}

			// Convert the recieved to string
			recString := string(buffer[0:rLen])
			//log.Info("Visuals received: ", recString)

			// Quickly check if it is an emergency
			if strings.Contains(recString, "!") {
				poiChan <- poiType{category: emergency}
				continue
			}

			// Otherwise we split it to seperate commands
			split := strings.Split(recString, "/")
			// That should be at least 3 long
			if len(split) < 3 {
				continue
			}
			// The first part of the command, is the type
			switch split[0] {
			case "r": //robot - recieve current position as 'r/x/y/r' - r is angle
				if tempX, err := strconv.Atoi(split[1]); err == nil {
					if tempY, err := strconv.Atoi(split[2]); err == nil {
						if tempR, err := strconv.Atoi(split[3]); err == nil {
							currentPos.x = tempX
							currentPos.y = tempY
							if tempR < -90 {
								currentPos.angle = tempR + 270
							} else {
								currentPos.angle = tempR - 90
							}
							//log.Info("Updated currentpos: ", currentPos)
						}
					}
				}
			case "b": //ball - recieve current position as 'b/x/y'
				if split[1] == "r" { // reset
					ballBuffer = []pointType{}
					//log.Info("Visuals: reset ball buffer")
					continue
				} else if split[1] == "d" { // list done
					if checkForNewBalls(balls, ballBuffer) {
						balls = make([]pointType, len(ballBuffer))
						copy(balls, ballBuffer)
						commandChan <- "compute"
					}
					continue
				}
				ball := pointType{}
				if tempX, err := strconv.Atoi(split[1]); err == nil {
					if tempY, err := strconv.Atoi(split[2]); err == nil {
						ball.x = tempX
						ball.y = tempY
						ballBuffer = append(ballBuffer, ball)
					}
				}
			case "p": // pixel distance
				log.Info("PixelDist: ", split)
			case "g": // goal
				if tempX, err := strconv.Atoi(split[1]); err == nil {
					if tempY, err := strconv.Atoi(split[2]); err == nil {
						goalPos.x = tempX
						goalPos.y = tempY
					}
				}
			}
		}
		active = false
	}
}

// checkForNewBalls will compare two slices and see if there are new elements
func checkForNewBalls(old, recevied []pointType) bool {
	for _, n := range recevied {
		found := false
		for _, o := range old {
			if isClose(o, n) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

// compares 2 points to see if they are close to each other
func isClose(old, new pointType) bool {
	_, len := old.dist(new)
	return len < 5
}

// sort balls purely based on length to closest
func (currentPos pointType) sortBalls(balls []pointType) (sortedBalls []pointType, err error) {
	origLength := len(balls)
	if origLength < 2 {
		sortedBalls = balls
		err = fmt.Errorf("Only %d balls, can't sort", origLength)
		return
	}
	//fmt.Println(currentPos.findRoute(balls, []pointType{}))

	for i := 0; i < origLength; i++ {
		minDist := 99999
		minI := 0
		for j, v := range balls {
			_, len := currentPos.dist(v)
			if len < minDist {
				minDist = len
				minI = j
			}
		}
		sortedBalls = append(sortedBalls, balls[minI])
		currentPos = balls[minI]
		balls = remove(balls, minI)
	}
	return
}

// remove an element from a slice
func remove(slice []pointType, s int) []pointType {
	return append(slice[:s], slice[s+1:]...)
}

// for getting and saving images - is not in use
func imageReciever() {
	udpServer, err := net.ListenPacket("udp", fmt.Sprintf(":%d", visualPort-1))
	if err != nil {
		log.Fatal(err)
	}
	defer udpServer.Close()

	buf := make([]byte, 16384)
	for {
		_, _, err := udpServer.ReadFrom(buf)
		if err != nil {
			continue
		}
		err = os.WriteFile("UI/Front/img/thumbnail.jpg", buf, 0644)
		if err != nil {
			continue
		}
	}
}
