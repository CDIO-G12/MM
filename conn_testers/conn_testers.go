package conntesters

import (
	u "MM/utils"
	"fmt"
	"net"
	"time"

	"github.com/jpillora/backoff"
	log "github.com/s00500/env_logger"
)

func VisualsClient(input <-chan string) {
	reconnTimer := &backoff.Backoff{
		Min:    500 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	log.Info("Running visuals client")
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", u.IP, u.VisualPort))
		if err != nil {
			time.Sleep(reconnTimer.Duration())
			continue
		}
		log.Info("Got visuals connection: ", conn.RemoteAddr().String())

		for {
			in := <-input
			_, err := conn.Write([]byte(in + "\n"))
			if err != nil {
				break
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func RobotClient(input <-chan string, output chan<- string) {
	reconnTimer := &backoff.Backoff{
		Min:    500 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	log.Info("Running robot client")
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", u.IP, u.RobotPort))
		if err != nil {
			time.Sleep(reconnTimer.Duration())
			continue
		}
		log.Info("Got robot connection: ", conn.RemoteAddr().String())

		go func() {
			buffer := make([]byte, 255)
			for {
				len, err := conn.Read(buffer)
				if err != nil {
					break
				}
				output <- string(buffer[0:len])
			}
		}()

		for {
			in := <-input
			_, err := conn.Write([]byte(in))
			if err != nil {
				break
			}
			time.Sleep(time.Millisecond)
		}
	}
}
