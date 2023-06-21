package main

import (
	f "MM/frame"
	l "MM/log"
	u "MM/utils"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/s00500/env_logger"
)

type states int

const (
	stateWait states = iota
	stateMoving
	stateExit
	stateNextPosQueue
	stateNextPos
	stateNextMove
	statePickup
	stateDump
	stateEmergency
	stateCalibrate
)

var state = stateWait

var stateMU = sync.RWMutex{}
var robLog l.Log_type

// initRobotServer is the main function for the robot server. In here are multiple goroutines and a statemachine to handle robot control.
func initRobotServer(frame *f.FrameType, keyChan <-chan string, poiChan <-chan u.PoiType, commandChan chan<- string) {
	addr := fmt.Sprintf("%s:%d", u.IP, u.RobotPort)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	robLog = l.Init_log("Robot", u.RobotPort-1)
	robLog.Log("Robot log connected")

	log.Infoln("Robot server is running on:", addr)
	for {
		// Waiting for an incomming client
		conn, err := server.Accept()
		if err != nil {
			log.Println("Failed to accept conn.", err)
			continue
		}
		log.Infoln("Connected to robot at:", conn.RemoteAddr().String())

		nextPos := u.PoiType{}
		ballCounter := 0
		moving := false
		newBallWaiting := false
		lastPickedBall := u.PointType{}

		//This helps for ending routines
		ctx, cancel := context.WithCancel(context.Background())

		//go routine that treats incomming POI
		go func() {
		loop:
			for {
				select {
				case poi := <-poiChan: // Incomming point of interest
					switch poi.Category { // Sorted by category
					case u.Emergency: // If emergency, we stop the robot
						_, err = conn.Write([]byte("!"))
						if err != nil {
							log.Println("Lost connection")
							break loop
						}
						setState(stateEmergency)

					case u.Start:
						if getState() == stateEmergency {
							setState(stateNextMove)
						}

					case u.Calibrate:
						setState(stateCalibrate)

					case u.Ball, u.Goal: //goal or ball
						if nextPos == poi && poi.Category != u.Goal {
							//setState(stateNextMove)
							continue
						}

						nextPos = poi
						log.Infoln("Got next pos: ", poi)
						setState(stateNextPosQueue)

					case u.NewBall:
						newBallWaiting = true

					case u.Found:
						setState(statePickup)

					case u.NotFound:
						if !moving {
							commandChan <- "next"
						} else {
							newBallWaiting = true
						}

					default:
						log.Infoln("Recieved weird POI? - ", poi)
					}

				case <-ctx.Done(): // If another routine is closed, this will end this routine
					break loop

				case key := <-keyChan: // Only used for manual control
					log.Infoln(manControlInt(key))
					_, err = conn.Write(manControlInt(key))
					if err != nil {
						log.Println("Lost connection")
						break loop
					}
				}
			}
			log.Infoln("LOOP BROKE!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			conn.Close()
		}()

		// This routine handles the incomming commands from the robot
		go func() {
			buffer := make([]byte, 32)
			for {
				// Read blocks until an incomming message comes, or the connection dies.
				len, err := conn.Read(buffer)
				if err != nil {
					log.Warnln("Lost connection to robot")
					break
				}
				// We convert from []byte to string
				recieved := string(buffer[0:len])
				robLog.Info("Got from robot: ", recieved)

				// Check what kind of command is recieved, and handle them
				switch {
				case strings.HasPrefix(recieved, "rd"): // ready - the initial command send, when the robot is ready to move
					if u.AutoGo {
						commandChan <- "go" // Send the first ball
					}

					log.Infoln("Robot ready!")
				case strings.HasPrefix(recieved, "gb"): // got ball - is send when the robot got a ball
					// Count the ball, and keep track of how many balls are stored at the moment
					ballCounter++
					log.Info("Ballcount: ", ballCounter)
					log.Info("Got Ball - time left: ", Timer.Left())

					// If the storage is full, we move to goal, otherwise we ask for the next ball
					if ballCounter >= u.BallCounterMax {
						commandChan <- "goal"
						continue
					}
					commandChan <- "gotBall" // got balls

				case strings.HasPrefix(recieved, "pb"): // pickedup ball
					commandChan <- "next"

				case strings.HasPrefix(recieved, "fm"): // finished move - is sent when the robot has done a move, and is waiting for next instruction
					// Every time we are done with a move, we ask for the current position, and runs the next move
					time.Sleep(u.TimeBetweenMovesMS * time.Millisecond)
					moving = false
					if getState() != stateEmergency {
						setState(stateNextMove)
					}
				case strings.HasPrefix(recieved, "m"):
					moving = true

				case strings.HasPrefix(recieved, "fd"): // finish dump
					ballCounter = 0
					log.Infoln("Dumped at: ", Timer.Now())
					commandChan <- "next"
				}
			}
			// if the loop is breaked, we cancel to stop the other routines
			cancel()
			setState(stateExit)
		}()

		// This handles the state machine for the robot
		positionQueue := []u.PoiType{}
		nextGoto := u.PoiType{}

		//state machine
	loop:
		for {
			// main state machine
			switch getState() {
			case stateExit: // exit is sent to exit the state machine
				commandChan <- "gone"
				break loop

			case stateWait, stateMoving: // Waits for a new state
				time.Sleep(100 * time.Millisecond)

			case stateNextPosQueue: // Create new array of moves
				if nextPos.Category == u.Ball {
					frame.RateBall(&nextPos.Point)
					log.Infoln("Rated ball as: ", nextPos.Point.Angle)
				}
				commandChan <- fmt.Sprintf("%d/%d/200/200/200\n", nextPos.Point.X, nextPos.Point.Y)
				positionQueue = frame.CreateMoves(u.CurrentPos.Get(), nextPos)
				if len(positionQueue) == 0 {
					fmt.Println("Houston. We have a problem.")
				}
				log.Info("Got position queue", positionQueue)
				send := ""
				for i, pos := range positionQueue {
					send += fmt.Sprintf("%d/%d/%d/%d/%d\n", pos.Point.X, pos.Point.Y, (100 + 25*i), (100 + 25*i), (100 + 25*i))
				}
				commandChan <- send
				setState(stateNextPos)

			case stateNextPos:
				if len(positionQueue) == 0 {
					log.Info("Should do something here!! ", nextPos, u.CurrentPos.Get())
					commandChan <- "next"
					setState(stateWait)
					continue
				}
				nextGoto = u.Pop(&positionQueue)
				if !frame.WithinBorder(nextGoto.Point) && false {
					log.Error("Error! Point outside of field", nextGoto)
					continue
				}
				setState(stateNextMove)

			case stateNextMove: // nextMove calculated the next move and sends it to the robot
				if nextGoto.Point.X == 0 && nextGoto.Point.Y == 0 {
					setState(stateNextPos)
					continue
				}
				if lastPickedBall == nextGoto.Point {
					commandChan <- "next"
					continue
				}

				// check if new ball have appeared on the way to goal
				if nextPos.Category == u.Goal && ballCounter < u.BallCounterMax && Timer.Left().Seconds() > 30 {
					commandChan <- "nextIf"
				}

				if newBallWaiting {
					newBallWaiting = false
					commandChan <- "next"
					setState(stateWait)
					continue
				}

				currentPos := u.CurrentPos.Get()
				angle, dist := currentPos.AngleAndDist(nextGoto.Point)
				dist = int(float64(dist) / u.GetPixelDist()) //convert from pixel to mm
				if nextGoto.Category == u.Ball {
					if nextGoto.Point.Angle >= u.RatingMiddleX {
						dist -= u.DistanceFromBallMiddleX
					} else if nextGoto.Point.Angle >= u.RatingCorner {
						dist -= u.DistanceFromBallCorner
					} else {
						dist -= u.DistanceFromBall
					}
				}

				// check if the move still makes sense
				if dist < 30 && nextPos.Category == u.Goal {
					commandChan <- "check"
				}

				log.Infof("Dist: %d, angle: %d, send angle: %d, next: %+v, current: %+v", dist, angle, angle-currentPos.Angle, nextGoto, currentPos)

				commandChan <- fmt.Sprintf("%d/%d/255/200/0\n", nextGoto.Point.X, nextGoto.Point.Y)

				if (dist < 75 && nextGoto.Category == u.WayPoint) || (dist < 30 && nextGoto.Category == u.PreciseWayPoint) {
					setState(stateNextPos)
					continue
					// if the angle is not very close to the current angle, or the robot is further away while the angle is not sort of correct, we send a rotation command
				}
				if dist < 30 && nextGoto.Category == u.Goal {
					setState(stateDump)
					continue
				}
				//fmt.Println(AbsRotation(angle - currentPos.Angle))
				if (AbsRotation(angle-currentPos.Angle) > 120) && dist < 200 && dist > 10 {
					success := sendToBot(conn, []byte{[]byte("B")[0], byte(dist + 5)})
					if !success {
						break loop
					}
				} else if (angle < currentPos.Angle-3 || angle > currentPos.Angle+3) && u.Abs(dist) > 200 {
					success := sendToBot(conn, calcRotation(angle-currentPos.Angle))
					if !success {
						break loop
					}
				} else if (nextGoto.Category == u.Ball || nextGoto.Category == u.Goal) && (angle < currentPos.Angle-3 || angle > currentPos.Angle+3) {
					success := sendToBot(conn, calcRotation(angle-currentPos.Angle))
					if !success {
						break loop
					}
				} else if angle < currentPos.Angle-5 || angle > currentPos.Angle+5 {
					success := sendToBot(conn, calcRotation(angle-currentPos.Angle))
					if !success {
						break loop
					}

					// if the distance is far away, we send a course forward
				} else if dist > 200 {
					dist /= 10
					if dist > 255 {
						dist = 255
					}
					success := sendToBot(conn, []byte{[]byte("F")[0], byte(int(dist - 2))})
					if !success {
						break loop
					}
				} else if dist < -3 {
					success := sendToBot(conn, []byte{[]byte("B")[0], byte(-dist + 10)})
					if !success {
						break loop
					}
					// if the distance is close, we send a fine forward
				} else if dist > 100 {
					success := sendToBot(conn, []byte{[]byte("f")[0], byte(dist - 20)})
					if !success {
						break loop
					}
				} else if dist > 10 {
					success := sendToBot(conn, []byte{[]byte("f")[0], byte(dist - 1)})
					if !success {
						break loop
					}
					// if we are very close to the wanted position we do something at some point
				} else {
					log.Infoln("Closing in!")
					if nextGoto.Category == u.WayPoint {
						setState(stateNextPos)
					} else if nextGoto.Category == u.Ball {
						commandChan <- fmt.Sprintf("check/%d/%d", nextGoto.Point.X, nextGoto.Point.Y)
						setState(stateWait)
					} else if nextGoto.Category == u.Goal {
						setState(stateDump)
					} else {
						log.Infoln("This is akward...", nextGoto)
						time.Sleep(time.Second)
					}
					continue
				}
				// set the new state to moving
				setState(stateMoving)

			case statePickup:
				commandChan <- "pause"
				lastPickedBall = nextGoto.Point
				success := sendToBot(conn, []byte{[]byte("T")[0], byte(nextGoto.Point.Angle / 10)})
				if !success {
					break loop
				}
				setState(stateWait)

			case stateDump:
				current := u.CurrentPos.Get()
				success := sendToBot(conn, calcRotation(0-current.Angle))
				if !success {
					break loop
				}
				time.Sleep(2 * time.Second)
				success = sendToBot(conn, []byte{[]byte("D")[0], byte(2)})
				if !success {
					break loop
				}
				setState(stateWait)

			case stateCalibrate:
				success := sendToBot(conn, []byte{[]byte("C")[0], byte(0)})
				if !success {
					break loop
				}
				time.Sleep(4 * time.Second)
				setState(stateNextPosQueue)

			case stateEmergency:
				log.Println("Emergency!")
				time.Sleep(1 * time.Second)
				success := sendToBot(conn, []byte{[]byte("B")[0], byte(50)})
				if !success {
					break loop
				}
				setState(stateWait)
			}
		}
	}
}

func getState() states {
	stateMU.RLock()
	defer stateMU.RUnlock()
	return state
}

func setState(newState states) {
	stateMU.Lock()
	defer stateMU.Unlock()
	state = newState
	robLog.Log("Updated state: ", state)
}

// sendToBot is used to send a certain package and returns a bool of success
func sendToBot(conn net.Conn, pkg []byte) bool {
	if len(pkg) > 1 {
		robLog.Log("Send ", string(pkg[0]), pkg[1], " to robot")
	} else {
		robLog.Log("Send ", string(pkg), " to robot")
	}
	_, err := conn.Write(pkg)
	if err != nil {
		log.Println("Lost connection")
		return false
	}
	//log.Debug("Send ", string(pkg[0]), pkg[1], " to robot")
	return true
}

func manControlInt(input string) (out []byte) {
	out = append(out, byte(input[0]))
	if arg, err := strconv.Atoi(input[1:]); err == nil {
		out = append(out, byte(arg))
	}
	return
}

// calcRotation checks if the rotation is clockwise or counter clockwise, and returns the command with the angle as a argument
func calcRotation(angle int) []byte {
	left := "L"
	right := "R"

	if angle > 180 {
		return []byte{[]byte(left)[0], byte(180 - (angle - 180))}
	} else if angle < -180 {
		return []byte{[]byte(right)[0], byte(180 + (angle + 180))}
	} else if angle < 0 { //if angle < 0 {
		return []byte{[]byte(left)[0], byte(-angle)}
	} else if angle > 0 {
		return []byte{[]byte(right)[0], byte(angle)}
	}
	return nil
}

func AbsRotation(angle int) int {
	if angle > 180 {
		return 180 - (angle - 180)
	} else if angle < -180 {
		return 180 + (angle + 180)
	} else if angle < 0 { //if angle < 0 {
		return angle
	} else if angle > 0 {
		return angle
	}
	return angle
}

func (s states) String() string {
	switch s {
	case stateDump:
		return "dump"
	case stateEmergency:
		return "emergency"
	case stateExit:
		return "exit"
	case stateMoving:
		return "moving"
	case stateNextMove:
		return "next move"
	case stateNextPos:
		return "next pos"
	case stateNextPosQueue:
		return "next pos queue"
	case statePickup:
		return "pickup"
	case stateWait:
		return "wait"
	}
	return "state"
}
