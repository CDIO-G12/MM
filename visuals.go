package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	f "MM/frame"
	u "MM/utils"

	log "github.com/s00500/env_logger"
)

var IsValid = regexp.MustCompile(`^[0-9\/gc]+$`).MatchString

// initVisualServer hold the visual server and handles stuff
func initVisualServer(frame *f.FrameType, poiChan chan<- u.PoiType, framePoiChan chan<- u.PoiType, commandChan chan string) {
	log.Info("Visual server started")
	//go imageReciever()
	balls := []u.PointType{}
	sortedBalls := []u.PointType{}
	currentPos := u.PointType{X: 200, Y: 200, Angle: 180}
	orangeBall := u.PointType{X: 0, Y: 0}
	goalPos := u.PointType{X: 200, Y: 400}
	//active := false
	//robotActive := false
	//firstSend := false
	lastCorners := [4]u.PointType{}
	sendChan := make(chan string, 5)

	// this routine handles commands comming from the robot
	go func() {
		for {
			// if it is not active, it will ignore commands from the robot
			/*if !active {
				time.Sleep(100 * time.Millisecond)
				continue
			}*/

			// recieve command from robot, and handle
			cmd := <-commandChan
			//fmt.Println("Recieved: ", cmd)
			switch cmd {

			case "first": // Send first ball
				log.Infoln("First ball send")

				if orangeBall.X != 0 {
					poiChan <- u.PoiType{Point: orangeBall, Category: u.Ball}
					continue
				}
				if len(sortedBalls) > 0 {
					poiChan <- u.PoiType{Point: sortedBalls[0], Category: u.Ball}
					//firstSend = true
				}

			case "next": // Send next ball
				if orangeBall.X != 0 {
					poiChan <- u.PoiType{Point: orangeBall, Category: u.Ball}
					continue
				}
				if len(sortedBalls) == 0 {
					poiChan <- u.PoiType{Point: goalPos, Category: u.Goal}
					log.Infoln("Goal send")
					continue
				} else {
					poiChan <- u.PoiType{Point: sortedBalls[0], Category: u.Ball}
				}

			case "pos": // Send current position
				poiChan <- u.PoiType{Point: currentPos, Category: u.Robot}

			case "goal":
				poiChan <- u.PoiType{Point: goalPos, Category: u.Goal}

			default:
				sendChan <- cmd
			}
		}
	}()

	// create server
	addr := fmt.Sprintf("%s:%d", u.IP, u.VisualPort)
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
		ballBuffer := []u.PointType{}
		//active = true

		go func() {
			for {
				data := <-sendChan
				if data == "exit" {
					break
				}
				_, err := conn.Write([]byte(data))
				if log.Should(err) {
					break
				}
			}
		}()

		for {
			// Read blocks, so we wait for incoming command
			rLen, err := conn.Read(buffer)
			// if it fails, we break the loop
			if log.Should(err) {
				// Send an emergency in a new routine so it wont block
				go func() {
					poiChan <- u.PoiType{Category: u.Emergency} // Stop the bot if connection is lost
				}()
				conn.Close()
				break
			}

			// Convert the recieved to string
			recString := string(buffer[0:rLen])
			outerSplit := strings.Split(recString, "\n")

			for _, recString := range outerSplit {
				if u.VisualDebugLog {
					log.Info("Visuals received: ", recString)
				}

				// Quickly check if it is an emergency
				if strings.Contains(recString, "!") {
					poiChan <- u.PoiType{Category: u.Emergency}
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
				case "o": //orange ball - recieve position as 'o/x/y'
					if tempX, err := strconv.Atoi(split[1]); err == nil {
						if tempY, err := strconv.Atoi(split[2]); err == nil {
							orangeBall.X = tempX
							orangeBall.Y = tempY
						}
					}

				case "r": //robot - recieve current position as 'r/x/y/r' - r is angle
					if tempX, err := strconv.Atoi(split[1]); err == nil && len(split) > 3 {
						if tempY, err := strconv.Atoi(split[2]); err == nil {
							if tempR, err := strconv.Atoi(split[3]); err == nil {
								currentPos.X = tempX
								currentPos.Y = tempY
								/*if tempR < -90 {
									currentPos.Angle = tempR + 270
								} else {
									currentPos.Angle = tempR - 90
								}*/
								currentPos.Angle = u.DegreeAdd(tempR, -90)
								//log.Info("Updated currentpos: ", currentPos)
								framePoiChan <- u.PoiType{Point: currentPos, Category: u.Robot}
								//poiChan <- u.PoiType{Point: currentPos, Category: u.Robot}
							}
						}
					}

				case "b": //ball - recieve current position as 'b/x/y'
					if split[1] == "r" { // reset
						ballBuffer = []u.PointType{}
						//log.Info("Visuals: reset ball buffer")
						continue
					} else if split[1] == "d" { // list done
						if checkForNewBalls(balls, ballBuffer) {
							balls = make([]u.PointType, len(ballBuffer))
							copy(balls, ballBuffer)

							var err error
							sortedBalls, err = currentPos.SortBalls(ballBuffer)
							if err != nil {
								return
							}
							log.Info("Computed balls: ", sortedBalls)
						}
						continue
					}
					ball := u.PointType{}
					if tempX, err := strconv.Atoi(split[1]); err == nil {
						if tempY, err := strconv.Atoi(split[2]); err == nil {
							ball.X = tempX
							ball.Y = tempY
							ballBuffer = append(ballBuffer, ball)
						}
					}

				case "c": // corner
					if tempI, err := strconv.Atoi(split[1]); err == nil && tempI < 4 && len(split) > 3 {
						if tempX, err := strconv.Atoi(split[2]); err == nil {
							if tempY, err := strconv.Atoi(split[3]); err == nil {
								corner := u.PointType{X: tempX, Y: tempY, Angle: tempI}
								if lastCorners[tempI] != corner {
									lastCorners[tempI] = corner
									framePoiChan <- u.PoiType{Category: u.Corner, Point: corner}

									if tempI == 3 {
										guide := frame.GetGuideFrame()
										send := ""
										for i, v := range guide {
											send += fmt.Sprintf("gc/%d/%d/%d\n", i, v.X, v.Y)
										}
										sendChan <- send
									}
								}
							}
						}
					}
				case "m": // middle x
					if tempI, err := strconv.Atoi(split[1]); err == nil && len(split) > 3 {
						if tempX, err := strconv.Atoi(split[2]); err == nil {
							if tempY, err := strconv.Atoi(split[3]); err == nil {
								poiChan <- u.PoiType{Category: u.MiddleXcorner, Point: u.PointType{X: tempX, Y: tempY, Angle: tempI}}
							}
						}
					}

				case "p": // pixel distance
					//log.Info("PixelDist: ", split)
					if s, err := strconv.ParseFloat(split[2], 32); err == nil {
						u.SetPixelDist(s)
					}

				case "g": // goal
					if tempX, err := strconv.Atoi(split[1]); err == nil {
						if tempY, err := strconv.Atoi(split[2]); err == nil {
							goalPos.X = tempX
							goalPos.Y = tempY
							framePoiChan <- u.PoiType{Point: goalPos, Category: u.Goal}
						}
					}
				}
			}
		}
		sendChan <- "exit"
		//active = false
	}
}

// checkForNewBalls will compare two slices and see if there are new elements
func checkForNewBalls(old, recevied []u.PointType) bool {
	if len(old) != len(recevied) {
		return true
	}
	for _, n := range recevied {
		found := false
		for _, o := range old {
			if u.IsClose(o, n, 5) {
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

// for getting and saving images - is not in use
func imageReciever() {
	udpServer, err := net.ListenPacket("udp", fmt.Sprintf(":%d", u.VisualPort-1))
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
