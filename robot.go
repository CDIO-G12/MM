package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/s00500/env_logger"
)

var lastKey string
var currentKey string
var simpleKeys = []string{"u", "d", "p", "b"}
var keyTranslator = map[string]string{
	"left":  "L",
	"right": "R",
	"up":    "F",
	"down":  "B",
}

func initRobotServer(keyChan <-chan string, poiChan <-chan poiType, commandChan chan<- string) {
	addr := fmt.Sprintf("%s:9999", ip)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	log.Println("Robot server is running on:", addr)
	stateChangeChan := make(chan string)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("Failed to accept conn.", err)
			continue
		}
		log.Infoln("Connected to robot at:", conn.RemoteAddr().String())

		currentPos := pointType{}
		nextPos := poiType{}
		ballCounter := 0

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
		loop:
			for {
				defer conn.Close()

				select {
				case <-ctx.Done():
					break loop

				case key := <-keyChan:
					log.Infoln(key)
					_, err = conn.Write(keyIntepreter(key))
					if err != nil {
						log.Println("Lost connection")
						break loop
					}

				case poi := <-poiChan:
					switch poi.category {
					case robot:
						currentPos = poi.point
						//log.Infoln("Updated current position: ", currentPos)

					case emergency:
						_, err = conn.Write([]byte("!"))
						if err != nil {
							log.Println("Lost connection")
							break loop
						}

					default: //goal or ball
						nextPos = poi
						log.Infoln("Got next pos: ", poi)
						stateChangeChan <- "nextMove"
					}
				}
			}
		}()

		go func() {
			lastBall := time.Now()
			buffer := make([]byte, 128)
			for {
				len, err := conn.Read(buffer)
				if err != nil {
					log.Println("Lost connection")
					break
				}
				rec := string(buffer[0:len])
				log.Println("Got from robot: ", rec)

				switch {
				case strings.Contains(rec, "rd"): //ready
					commandChan <- "pos"
					commandChan <- "first"

				case strings.Contains(rec, "gb"): //got ball
					if time.Since(lastBall) < time.Second*2 {
						continue
					}
					lastBall = time.Now()
					fmt.Println("GOT BALL")
					ballCounter++
					if ballCounter >= 3 {
						ballCounter = 0
						commandChan <- "goal"
					} else {
						commandChan <- "next"
					}
					fallthrough
				case strings.Contains(rec, "fm"): //finished move
					commandChan <- "pos"
					time.Sleep(time.Second / 2)
					stateChangeChan <- "nextMove"
				}
			}
			cancel()
		}()

		state := "waiting"

		go func() {
			for {
				select {
				case <-ctx.Done():
					state = "exit"
					break

				case state = <-stateChangeChan:
					log.Info("New state: ", state)
				}
			}
		}()

	loop:
		for {
			switch state {
			case "exit":
				break loop
			case "waiting":
				time.Sleep(100 * time.Millisecond)

			case "moving":
				time.Sleep(100 * time.Millisecond)

			case "nextMove":
				angle, dist := currentPos.dist(nextPos.point)
				log.Infof("Dist %d, angle %d, next %+v, current %+v", dist, angle, nextPos.point, currentPos)
				if (angle < currentPos.angle-5 || angle > currentPos.angle+5) || ((angle < currentPos.angle-15 || angle > currentPos.angle+15) && dist > 100) {
					success := sendToBot(conn, calcRotation((currentPos.angle - angle)))
					if !success {
						break loop
					}
				} else if dist > 255 {
					success := sendToBot(conn, []byte{[]byte("F")[0], byte(dist / 255)})
					if !success {
						break loop
					}
				} else if dist > 3 {
					success := sendToBot(conn, []byte{[]byte("f")[0], byte(dist)})
					if !success {
						break loop
					}
				} else {
					//TODO: do something here!
				}

				state = "moving"
			}
		}
	}
}

func sendToBot(conn net.Conn, pkg []byte) bool {

	_, err := conn.Write(pkg)
	if err != nil {
		log.Println("Lost connection")
		return false
	}
	//log.Debug("Send ", string(pkg[0]), pkg[1], " to robot")
	return true
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

/*
up - send up
left - send left
\left - send up
\up - send stop



*/

func calcRotation(angle int) []byte {
	if angle > 0 {
		return []byte{[]byte("L")[0], byte(angle)}
	} else if angle < 0 {
		return []byte{[]byte("R")[0], byte(-angle)}
	}
	return nil
}
