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

//var balls = []pointType{{x: 2, y: 3}, {x: 4, y: 2}, {x: 7, y: 3}, {x: 5, y: 4}, {x: 3, y: 6}, {x: 8, y: 6}, {x: 5, y: 8}, {x: 7, y: 8}, {x: 2, y: 9}, {x: 10, y: 10}}

func initVisualServer(poiChan chan<- poiType, commandChan chan string) {
	log.Info("Visual server started")
	//go imageReciever()
	balls := []pointType{}
	sortedBalls := []pointType{}
	currentPos := pointType{x: 200, y: 200, angle: 180}
	goalPos := pointType{x: 200, y: 400}
	active := false

	go func() {
		for {
			if !active {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			cmd := <-commandChan
			switch cmd {
			case "first": // Send first ball
				log.Infoln("First ball send")
				if len(sortedBalls) > 0 {
					poiChan <- poiType{point: sortedBalls[0], category: ball}
				}
			case "next": // Send next ball
				if len(sortedBalls) == 0 {
					poiChan <- poiType{point: goalPos, category: goal}
					log.Infoln("Goal send")
				} else {
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

	//commandChan <- "compute"

	addr := fmt.Sprintf("%s:%d", ip, visualPort)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	log.Println("Visuals server is running on:", addr)

	for {
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
			rLen, err := conn.Read(buffer)
			if log.Should(err) {
				go func() {
					poiChan <- poiType{category: emergency} // Stop the bot if connection is lost
				}()
				conn.Close()
				break
			}
			recString := string(buffer[0:rLen])
			//log.Info("Visuals received: ", recString)
			if strings.Contains(recString, "!") {
				poiChan <- poiType{category: emergency}
				continue
			}

			split := strings.Split(recString, "/")
			if len(split) < 3 {
				continue
			}
			switch split[0] {
			case "r": //robot - recieve current position as 'r/x/y/z' - z is angle
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
			case "p": //pixel distance
				log.Info("PixelDist: ", split)
			case "g":
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

func remove(slice []pointType, s int) []pointType {
	return append(slice[:s], slice[s+1:]...)
}

// for getting and saving images
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
