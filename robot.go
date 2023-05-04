package main

import (
	f "MM/frame"
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
)

var state = stateWait
var stateMU = sync.Mutex{}

// initRobotServer is the main function for the robot server. In here are multiple goroutines and a statemachine to handle robot control.
func initRobotServer(frame *f.FrameType, keyChan <-chan string, poiChan <-chan u.PoiType, commandChan chan<- string) {
	addr := fmt.Sprintf("%s:%d", u.IP, u.RobotPort)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	log.Infoln("Robot server is running on:", addr)
	for {
		// Waiting for an incomming client
		conn, err := server.Accept()
		if err != nil {
			log.Println("Failed to accept conn.", err)
			continue
		}
		log.Infoln("Connected to robot at:", conn.RemoteAddr().String())

		currentPos := u.SafePointType{}

		nextPos := u.PoiType{}
		ballCounter := 0

		//This helps for ending routines
		ctx, cancel := context.WithCancel(context.Background())

		//go routine that treats incomming POI
		go func() {
		loop:
			for {
				select {
				case poi := <-poiChan: // Incomming point of interest
					//log.Infoln("Recieved POI", poi)
					switch poi.Category { // Sorted by category
					case u.Robot:
						currentPos.Set(poi.Point)
						//log.Infoln("Updated current position: ", poi.Point)
						continue

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

					case u.Ball, u.Goal: //goal or ball
						nextPos = poi
						log.Infoln("Got next pos: ", poi)
						setState(stateNextPosQueue)

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
				log.Infoln("Got from robot: ", recieved)

				// Check what kind of command is recieved, and handle them
				switch {
				case strings.Contains(recieved, "rd"): // ready - the initial command send, when the robot is ready to move
					go func() {
						commandChan <- "ready"
						time.Sleep(time.Millisecond)
						commandChan <- "pos" // Ask for current position
						time.Sleep(time.Millisecond)
						commandChan <- "next" // Send the first ball
						time.Sleep(time.Millisecond)
						log.Infoln("Robot ready!")
					}()

				case strings.Contains(recieved, "gb"): // got ball - is send when the robot got a ball
					// This might be send multiple types at once, so this will 'debounce' it

					// Count the ball, and keep track of how many balls are stored at the moment
					ballCounter++
					// If the storage is full, we move to goal, otherwise we ask for the next ball
					if ballCounter >= u.BallCounterMax {
						commandChan <- "goal"
					} else {
						commandChan <- "next"
					}
					// fallthrough
				case strings.Contains(recieved, "fm"): //finished move - is sent when the robot has done a move, and is waiting for next instruction
					// Every time we are done with a move, we ask for the current position, and runs the next move
					time.Sleep(time.Second / 7)
					commandChan <- "pos"
					time.Sleep(10 * time.Millisecond)
					setState(stateNextMove)

				case strings.Contains(recieved, "fd"):
					ballCounter = 0
					commandChan <- "next"

				}
			}
			// if the loop is breaked, we cancel to stop the other routines
			cancel()
			setState(stateExit)
		}()

		// This handles the state machine for the robot
		positions := []u.PoiType{}
		nextGoto := u.PoiType{}
	loop:
		for {
			// main state machine
			switch getState() {
			case stateExit: // exit is sent to exit the state machine
				commandChan <- "gone"
				break loop

			case stateWait: // Waits for a new state
				time.Sleep(100 * time.Millisecond)

			case stateMoving: // Moving waits for a '' command
				time.Sleep(100 * time.Millisecond)

			case stateNextPosQueue: // Create new array of moves
				commandChan <- "pos"
				time.Sleep(10 * time.Millisecond)
				positions = frame.CreateMoves(currentPos.Get(), nextPos)
				log.Infoln("Got position queue", positions)
				send := ""
				for i, pos := range positions {
					send += fmt.Sprintf("%d/%d/%d/%d/%d\n", pos.Point.X, pos.Point.Y, (100 + 25*i), (100 + 25*i), (100 + 25*i))
				}
				commandChan <- send
				setState(stateNextPos)

			case stateNextPos:
				commandChan <- "pos"
				if len(positions) == 0 {
					if _, temp := currentPos.Get().Dist(nextPos.Point); temp < 1000 {
						switch nextPos.Category {
						case u.Goal:
							//Dump
						case u.Ball:
							//PickupBall
						}
						log.Info("Should do something here!! ", nextPos, currentPos.Get())
						continue
					}
					setState(stateNextPosQueue)
					continue
				}
				nextGoto = u.Pop(&positions)
				if nextGoto.Point.X == 0 {
					log.Error("Error! Could not get next goto")
					continue
				}
				setState(stateNextMove)

			case stateNextMove: // nextMove calculated the next move and sends it to the robot
				commandChan <- "pos"
				time.Sleep(50 * time.Millisecond)

				if nextGoto.Point.X == 0 && nextGoto.Point.Y == 0 {
					setState(stateNextPos)
					continue
				}
				current := currentPos.Get()
				angle, dist := current.Dist(nextGoto.Point)
				if nextGoto.Category == u.Ball {
					dist = int(float64(dist)/u.GetPixelDist()) - 280 // convert pixel distance to distance in millimeter
				}

				log.Infof("Dist: %d, angle: %d, send angle: %d, next: %+v, current: %+v", dist, angle, angle-current.Angle, nextGoto, current)

				commandChan <- fmt.Sprintf("%d/%d/255/200/0\n", nextGoto.Point.X, nextGoto.Point.Y)

				// if the angle is not very close to the current angle, or the robot is further away while the angle is not sort of correct, we send a rotation command
				if (angle < current.Angle-2 || angle > current.Angle+2) || ((angle < current.Angle-20 || angle > current.Angle+20) && dist > 200) {
					success := sendToBot(conn, calcRotation(angle-current.Angle))
					if !success {
						break loop
					}
					// if the distance is far away, we send a course forward
				} else if dist > 255 {
					success := sendToBot(conn, []byte{[]byte("F")[0], byte(dist / 255)})
					if !success {
						break loop
					}
					// if the distance is close, we send a fine forward
				} else if dist < 0 {
					success := sendToBot(conn, []byte{[]byte("B")[0], byte(-dist)})
					if !success {
						break loop
					}
				} else if dist > 50 {
					success := sendToBot(conn, []byte{[]byte("f")[0], byte(dist)})
					if !success {
						break loop
					}
				} else if dist > 5 {
					success := sendToBot(conn, []byte{[]byte("f")[0], byte(dist)})
					if !success {
						break loop
					}
					// if we are very close to the wanted position we do something at some point
				} else {
					log.Infoln("Closing in!")
					if nextGoto.Category == u.WayPoint {
						setState(stateNextPos)
					} else if nextGoto.Category == u.Ball {
						setState(statePickup)
					} else if nextGoto.Category == u.Goal {
						setState(stateDump)
					} else {
						log.Infoln("This is akward...", nextGoto)
						time.Sleep(time.Second)
					}
					continue
				}
				// set the new state to moving
				log.Infoln("Moving!")
				setState(stateMoving)

			case statePickup:
				success := sendToBot(conn, []byte{[]byte("T")[0], 0})
				if !success {
					break loop
				}
				setState(stateWait)

			case stateDump:
				success := sendToBot(conn, calcRotation(0-currentPos.Get().Angle))
				if !success {
					break loop
				}
				time.Sleep(1 * time.Second)
				log.Infoln("GOOOOOOOOOOOOOOOOOOOOOAAAAALL!!!!")
				success = sendToBot(conn, []byte{[]byte("D")[0], byte(ballCounter)})
				if !success {
					break loop
				}
				setState(stateWait)

			case stateEmergency:
				log.Println("Emergency!")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func getState() states {
	stateMU.Lock()
	defer stateMU.Unlock()
	return state
}

func setState(newState states) {
	stateMU.Lock()
	defer stateMU.Unlock()
	state = newState
	log.Info("Updated state: ", state)
}

// sendToBot is used to send a certain package and returns a bool of success
func sendToBot(conn net.Conn, pkg []byte) bool {
	if len(pkg) > 1 {
		log.Infoln("Send ", string(pkg[0]), pkg[1], " to robot")
	} else {
		log.Infoln("Send to bot: ", string(pkg))
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
	if angle > 180 {
		return []byte{[]byte("L")[0], byte(180 - (angle - 180))}
	} else if angle < -180 {
		return []byte{[]byte("R")[0], byte(180 + (angle + 180))}
	} else if angle < 0 { //if angle < 0 {
		return []byte{[]byte("L")[0], byte(-angle)}
	} else if angle > 0 {
		return []byte{[]byte("R")[0], byte(angle)}
	}
	return nil
}
