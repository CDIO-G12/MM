package main

import (
	"net"
	"time"

	"github.com/jpillora/backoff"
	log "github.com/s00500/env_logger"
)

func keyGet(keyChan chan<- string) {
	reconnTimer := &backoff.Backoff{
		Min:    500 * time.Millisecond,
		Max:    3 * time.Second,
		Factor: 2,
		Jitter: false,
	}
	buffer := make([]byte, 128)
	log.Info("Dialing keygetter")
	for {
		conn, err := net.Dial("tcp", "localhost:10023")
		if err != nil {
			time.Sleep(reconnTimer.Duration())
			continue
		}
		log.Info("Connected to keygetter at: ", conn.RemoteAddr().String())
		for {
			len, err := conn.Read(buffer)
			if err != nil {
				log.Errorf("Keygetter: error reading: %v", err)
				keyChan <- "!"
				break
			}
			keyChan <- string(buffer[0:len])
		}

	}
}
