package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	f "MM/frame"
	l "MM/log"
	u "MM/utils"

	log "github.com/s00500/env_logger"
)

// initVisualServer hold the visual server and handles stuff
func initVisualServer(frame *f.FrameType, poiChan chan<- u.PoiType, framePoiChan chan<- u.PoiType, commandChan chan string) {
	//go imageReciever()
	balls := []u.PointType{}
	sortedBalls := []u.PointType{}
	orangeBall := u.PointType{X: 0, Y: 0}
	goalPos := u.PointType{X: 48, Y: 355}
	lastBallRecieveTime := time.Now()
	currentBallLastSeen := time.Now()
	currentBall := u.PointType{Angle: -1}
	robotActive := false
	robotWaiting := false
	lastCorners := [4]u.PointType{}
	sendChan := make(chan string, 50)
	sendOrange := false

	visLog := l.Init_log("Visuals", u.VisualPort-1)
	visLog.Log("Visuals log connected")

	// this routine handles commands comming from the robot
	go func() {
		for {
			// recieve command from robot, and handle
			cmd := <-commandChan
			visLog.Info("Recieved: ", cmd)

			switch cmd {
			case "ready":
				robotActive = true

			case "gone", "pause":
				robotActive = false

			case "gotBall":
				if sendOrange {
					safeChannelSend(sendChan, "no\n")
				}

			case "nextIf":
				if len(sortedBalls) == 0 && orangeBall.X == 0 {
					continue
				}
				fallthrough

			case "next", "go": // Send next ball
				if cmd == "go" {
					Timer.Start()
				}
				robotActive = true

				if orangeBall.X != 0 {
					currentBall = orangeBall
					poiChan <- u.PoiType{Point: orangeBall, Category: u.Ball}
					continue
				}

				if len(sortedBalls) == 0 {
					currentBall.X = 0
					poiChan <- u.PoiType{Point: goalPos, Category: u.Goal}
					continue
				}

				ballBuffer := make([]u.PointType, len(sortedBalls))
				copy(ballBuffer, sortedBalls)
				sortedBalls = frame.SortBalls(ballBuffer)

				if len(sortedBalls) == 0 {
					currentBall.X = 0
					poiChan <- u.PoiType{Point: goalPos, Category: u.Goal}
				} else {
					// Dont send the same ball twice
					/*if currentBall.IsClose(sortedBalls[0], 20) {
						if len(sortedBalls) > 1 {
							currentBall = sortedBalls[1]
						} else {
							currentBall.X = 0
							poiChan <- u.PoiType{Point: goalPos, Category: u.Goal}
							continue
						}
					} else {*/
					currentBall = sortedBalls[0]
					//}
					poiChan <- u.PoiType{Point: currentBall, Category: u.Ball}
				}

			case "goal":
				currentBall.X = 0
				poiChan <- u.PoiType{Point: goalPos, Category: u.Goal}

			case "check":
				//check if current is not a ball
				if currentBall.X == 0 {
					continue
				}
				//if there is not much time left, we go to dump
				if Timer.Left().Seconds() < 30 && u.DumpAtThirty {
					safeChannelSend(commandChan, "goal")
					continue
				}

				// if we have not seen the current ball in a long time
				if time.Since(currentBallLastSeen).Seconds() > u.SecondsBeforeForgetBall && time.Since(lastBallRecieveTime).Seconds() < u.SecondsBeforeForgetBall {
					poiChan <- u.PoiType{Category: u.NewBall}
				}

			default:
				if strings.Contains(cmd, "check") {
					spl := strings.Split(cmd, "/")
					point := u.PointType{}
					point.X, _ = strconv.Atoi(spl[1])
					point.Y, _ = strconv.Atoi(spl[2])
					visLog.Info("Check: ", spl)

					/*
						if time.Since(currentBallLastSeen).Seconds() > u.SecondsBeforeForgetBall && time.Since(lastBallRecieveTime).Seconds() < u.SecondsBeforeForgetBall {
							poiChan <- u.PoiType{Category: u.NotFound}
						} else {
							poiChan <- u.PoiType{Category: u.Found}
						}
					*/

					if u.InArrayClose(point, sortedBalls) || point.IsClose(orangeBall, 15) {
						poiChan <- u.PoiType{Category: u.Found}
					} else {
						poiChan <- u.PoiType{Category: u.NotFound}
					}
					continue
				}

				safeChannelSend(sendChan, cmd)
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
		ballBuffer := []u.PointType{}
		//active = true

		if robotActive {
			poiChan <- u.PoiType{Category: u.Start}
		}

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
			//close(sendChan)
		}()

		buffer := make([]byte, 65536)
		for {
			// Read blocks, so we wait for incoming command
			rLen, err := conn.Read(buffer)
			// if it fails, we break the loop
			if log.Should(err) {
				// Send an emergency in a new routine so it wont block
				poiChan <- u.PoiType{Category: u.Emergency} // Stop the bot if connection is lost
				conn.Close()
				break
			}

			// Convert the recieved to string
			recString := string(buffer[0:rLen])
			outerSplit := strings.Split(recString, "\n")
			currentPosLocal := u.CurrentPos.Get()
			visLog.Log("Recieved: ", recString)

			for _, recString := range outerSplit {

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
				if u.VisualDebugLog {
					log.Info("Visuals received: ", recString, split)
				}

				// The first part of the command, is the type
				switch split[0] {
				case "o": //orange ball - recieve position as 'o/x/y'
					if tempX, err := strconv.Atoi(split[1]); err == nil {
						if tempY, err := strconv.Atoi(split[2]); err == nil {
							orangeBall.X = tempX
							orangeBall.Y = tempY
							if currentBall.IsClose(orangeBall, 25) {
								currentBallLastSeen = time.Now()
							}
						}
					}

				case "r": //robot - recieve current position as 'r/x/y/r' - r is angle
					if tempX, err := strconv.Atoi(split[1]); err == nil && len(split) > 3 {
						if tempY, err := strconv.Atoi(split[2]); err == nil {
							if tempR, err := strconv.Atoi(split[3]); err == nil {
								current := u.PointType{}
								current.X = tempX
								current.Y = tempY

								if current.X < 2 || current.Y < 2 {
									continue
								}
								current.Angle = u.DegreeAdd(tempR, -90)

								if !currentPosLocal.IsClose(current, 5) {
									visLog.Log("Updated currentpos: ", current)
								}
								u.CurrentPos.Set(current)

							}
						}
					}

				case "b": //ball - recieve current position as 'b/x/y'
					if split[1] == "r" { // reset
						ballBuffer = []u.PointType{}
						//log.Info("Visuals: reset ball buffer")
						continue
					}
					if split[1] == "d" { // list done
						if len(ballBuffer) <= 0 {
							sortedBalls = ballBuffer
							continue
						}
						lastBallRecieveTime = time.Now()
						if !checkForNewBalls(balls, ballBuffer) {
							continue
						}
						balls = make([]u.PointType, len(ballBuffer))
						copy(balls, ballBuffer)

						sortedBalls = frame.SortBalls(ballBuffer)
						if len(sortedBalls) < 1 {
							continue
						}

						visLog.Log("Computed balls: ", sortedBalls)

						for i, v := range sortedBalls {
							safeChannelSend(sendChan, fmt.Sprintf("b/%d/%d/%d\n", i+1, v.X, v.Y))
						}

						/*if !u.InArray(nextBall, sortedBalls) && !nextBall.IsClose(orangeBall, 4) {
							poiChan <- u.PoiType{Point: sortedBalls[0], Category: u.Ball}
							continue
						}*/

						if robotWaiting {
							robotWaiting = false
							poiChan <- u.PoiType{Point: sortedBalls[0], Category: u.Ball}
							continue
						}

						/*if nextBall != sortedBalls[0] && nextBall != orangeBall {
							poiChan <- u.PoiType{Point: sortedBalls[0], Category: u.Ball}
						}*/

						continue
					}

					ball := u.PointType{}
					if tempX, err := strconv.Atoi(split[1]); err == nil {
						if tempY, err := strconv.Atoi(split[2]); err == nil {
							ball.X = tempX
							ball.Y = tempY

							if currentBall.IsClose(ball, 5) {
								currentBallLastSeen = time.Now()
							}

							ang, dis := currentPosLocal.AngleAndDist(ball)
							if dis < 50 {
								if u.Abs(ang-currentPosLocal.Angle) < 15 {
									continue
								}
							}
							ballBuffer = append(ballBuffer, ball)
						}
					}

				case "c": // corner - c/i/x/y
					if tempI, err := strconv.Atoi(split[1]); err == nil && tempI < 4 && len(split) > 3 {
						if tempX, err := strconv.Atoi(split[2]); err == nil {
							if tempY, err := strconv.Atoi(split[3]); err == nil && tempY > 2 && tempX > 2 {
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
										safeChannelSend(sendChan, send)
									}
								}
							}
						}
					}

				case "m": // middle x
					if tempI, err := strconv.Atoi(split[1]); err == nil && tempI < 4 && len(split) > 3 {
						if tempX, err := strconv.Atoi(split[2]); err == nil {
							if tempY, err := strconv.Atoi(split[3]); err == nil && tempY > 2 && tempX > 2 {
								framePoiChan <- u.PoiType{Category: u.MiddleXcorner, Point: u.PointType{X: tempX, Y: tempY, Angle: tempI}}
							}
						}
					}

				case "p": // pixel distance
					//log.Info("PixelDist: ", split)
					if s, err := strconv.ParseFloat(split[2], 32); err == nil {
						u.SetPixelDist(s)
					}

				case "f": // found - not used at the moment
					/*if split[1] == "t" {
						poiChan <- u.PoiType{Category: u.Found}
					} else {
						poiChan <- u.PoiType{Category: u.NotFound}
					}*/

				case "g": // goal
					//continue
					if tempX, err := strconv.Atoi(split[1]); err == nil {
						if tempY, err := strconv.Atoi(split[2]); err == nil {
							tempX += int(u.MmToGoal * u.GetPixelDist())
							goalPos.X = tempX
							goalPos.Y = tempY
							safeChannelSend(framePoiChan, u.PoiType{Point: goalPos, Category: u.Goal})
						}
					}
				}
			}
		}
		safeChannelSend(sendChan, "exit\n")
		//active = false
	}
}

func safeChannelSend[t any](ch chan<- t, command t) {
	select {
	case ch <- command:
	default:
		log.Error("Channel full or closed on command: ", command)
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
			if o.IsClose(n, 3) {
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
