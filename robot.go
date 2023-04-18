package main

import (
	f "MM/frame"
	u "MM/utils"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/s00500/env_logger"
)

// initRobotServer is the main function for the robot server. In here are multiple goroutines and a statemachine to handle robot control.
func initRobotServer(frame *f.FrameType, keyChan <-chan string, poiChan <-chan u.PoiType, commandChan chan<- string, pixelDist *u.PixelDistType) {
	addr := fmt.Sprintf("%s:%d", u.IP, u.RobotPort)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	log.Infoln("Robot server is running on:", addr)
	stateMU := sync.RWMutex{}
	state := "waiting"
	for {
		// Waiting for an incomming client
		conn, err := server.Accept()
		if err != nil {
			log.Println("Failed to accept conn.", err)
			continue
		}
		log.Infoln("Connected to robot at:", conn.RemoteAddr().String())

		currentPos := u.PointType{}
		nextPos := u.PoiType{}
		ballCounter := 0

		//This helps for ending routines
		ctx, cancel := context.WithCancel(context.Background())

		//go routine that treats incomming POI
		go func() {
		loop:
			for {
				defer conn.Close()

				select {
				case <-ctx.Done(): // If another routine is closed, this will end this routine
					break loop

				case key := <-keyChan: // Only used for manual control
					log.Infoln(key)
					_, err = conn.Write(keyIntepreter(key))
					if err != nil {
						log.Println("Lost connection")
						break loop
					}

				case poi := <-poiChan: // Incomming point of interest
					switch poi.Category { // Sorted by category
					case u.Robot:
						currentPos = poi.Point
						//log.Infoln("Updated current position: ", currentPos)

					case u.Emergency: // If emergency, we stop the robot
						_, err = conn.Write([]byte("!"))
						if err != nil {
							log.Println("Lost connection")
							break loop
						}
						// Probably need to do more here

					default: //goal or ball
						nextPos = poi
						log.Infoln("Got next pos: ", poi)
						stateMU.Lock()
						state = "nextMove"
						stateMU.Unlock()
					}
				}
			}
		}()

		// This routine handles the incomming commands from the robot
		go func() {
			lastBall := time.Now()
			buffer := make([]byte, 32)
			for {
				// Read blocks until an incomming message comes, or the connection dies.
				len, err := conn.Read(buffer)
				if err != nil {
					log.Println("Lost connection")
					break
				}
				// We convert from []byte to string
				recieved := string(buffer[0:len])
				log.Println("Got from robot: ", recieved)

				// Check what kind of command is recieved, and handle them
				switch {
				case strings.Contains(recieved, "rd"): // ready - the initial command send, when the robot is ready to move
					commandChan <- "pos"   // Ask for current position
					commandChan <- "first" // Send the first ball

				case strings.Contains(recieved, "gb"): // got ball - is send when the robot got a ball
					// This might be send multiple types at once, so this will 'debounce' it
					if time.Since(lastBall) < time.Second*2 {
						continue
					}
					lastBall = time.Now()

					// Count the ball, and keep track of how many balls are stored at the moment
					ballCounter++
					// If the storage is full, we move to goal, otherwise we ask for the next ball
					if ballCounter >= u.BallCounterMax {
						ballCounter = 0
						commandChan <- "goal"
					} else {
						commandChan <- "next"
					}
					// fallthrough
				case strings.Contains(recieved, "fm"): //finished move - is sent when the robot has done a move, and is waiting for next instruction
					// Every time we are done with a move, we ask for the current position, and runs the next move
					commandChan <- "pos"
					time.Sleep(time.Second / 2)
					stateMU.Lock()
					state = "nextMove"
					stateMU.Unlock()
				}
			}
			// if the loop is breaked, we cancel to stop the other routines
			cancel()
			state = "exit"
		}()

		// This handles the state machine for the robot
		localState := ""
		positions := []u.PointType{}
		nextGoto := u.PointType{}
	loop:
		for {
			// critical read
			stateMU.RLock()
			localState = state
			stateMU.RUnlock()

			// main state machine
			switch localState {
			case "exit": // exit is sent to exit the state machine
				break loop
			case "waiting": // Waits for a new state
				time.Sleep(100 * time.Millisecond)

			case "moving": // Moving waits for a 'fm' command
				time.Sleep(100 * time.Millisecond)

			case "nextPosQueue": // Create new array of moves
				positions = frame.CreateMoves(currentPos, nextPos.Point)
				stateMU.Lock()
				state = "nextPos"
				stateMU.Unlock()

			case "nextPos":
				if len(positions) == 0 {
					if _, temp := currentPos.Dist(nextPos.Point); temp < 5 {
						switch nextPos.Category {
						case u.Goal:
							//Dump
						case u.Ball:
							//PickupBall
						}
						log.Info("Should do something here!! ", nextPos, currentPos)
						continue
					}

					stateMU.Lock()
					state = "nextPosQueue"
					stateMU.Unlock()
					continue
				}
				nextGoto = u.Pop(&positions)

			case "nextMove": // nextMove calculated the next move and sends it to the robot
				angle, dist := currentPos.Dist(nextGoto)
				log.Infof("Dist: %d, angle: %d, next: %+v, current: %+v", dist, angle, nextGoto, currentPos)
				// if the angle is not very close to the current angle, or the robot is further away while the angle is not sort of correct, we send a rotation command
				if (angle < currentPos.Angle-5 || angle > currentPos.Angle+5) || ((angle < currentPos.Angle-15 || angle > currentPos.Angle+15) && dist > 100) {
					success := sendToBot(conn, calcRotation((currentPos.Angle - angle)))
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
				} else if dist > 3 {
					success := sendToBot(conn, []byte{[]byte("f")[0], byte(dist)})
					if !success {
						break loop
					}
					// if we are very close to the wanted position we do something at some point
				} else {
					stateMU.Lock()
					state = "nextPos"
					stateMU.Unlock()
					continue
				}
				// set the new state to moving
				stateMU.Lock()
				state = "moving"
				stateMU.Unlock()

			case "emergency":
				log.Println("Emergency!")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

// sendToBot is used to send a certain package and returns a bool of success
func sendToBot(conn net.Conn, pkg []byte) bool {

	_, err := conn.Write(pkg)
	if err != nil {
		log.Println("Lost connection")
		return false
	}
	//log.Debug("Send ", string(pkg[0]), pkg[1], " to robot")
	return true
}

// deprecated - keyIntepreter is used for manual control

var lastKey string
var currentKey string
var simpleKeys = []string{"u", "d", "p", "b"}
var keyTranslator = map[string]string{
	"left":  "L",
	"right": "R",
	"up":    "F",
	"down":  "B",
}

func keyIntepreter(key string) []byte {
	remove := strings.HasPrefix(key, "\\")
	if remove {
		key = strings.Replace(key, "\\", "", 1)
	}

	for _, sk := range simpleKeys {
		if key != sk {
			continue
		}
		if !remove {
			return []byte(key)
		}
		return []byte{}
	}

	if strings.HasSuffix(key, "space") {
		if remove {
			return []byte("s")
		} else {
			return []byte("S")
		}
	}

	if remove {
		if currentKey == key {
			if lastKey != "" {
				currentKey = lastKey
				lastKey = ""
				return []byte(keyTranslator[currentKey])
			}
			currentKey = ""
			return []byte("!")
		} else if lastKey == key {
			lastKey = ""
			return []byte("")
		}
	}
	if currentKey != key {
		lastKey = currentKey
	}

	if key == "" {
		return []byte("!")
	}
	currentKey = key
	return []byte(keyTranslator[key])
}

// calcRotation checks if the rotation is clockwise or counter clockwise, and returns the command with the angle as a argument
func calcRotation(angle int) []byte {
	if angle > 0 {
		return []byte{[]byte("L")[0], byte(angle)}
	} else if angle < 0 {
		return []byte{[]byte("R")[0], byte(-angle)}
	}
	return nil
}
