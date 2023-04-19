package utils

import (
	"fmt"
	"net"
	"time"

	"github.com/jpillora/backoff"
	log "github.com/s00500/env_logger"
)

func KeyGet(keyChan chan<- string) {
	reconnTimer := &backoff.Backoff{
		Min:    500 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
		Jitter: false,
	}
	buffer := make([]byte, 128)
	log.Info("Dialing keygetter")
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", IP, keyGetterPort))
		if err != nil {
			time.Sleep(reconnTimer.Duration())
			continue
		}
		log.Info("Connected to keygetter at: ", conn.RemoteAddr().String())
		for {
			len, err := conn.Read(buffer)
			if err != nil {
				log.Errorf("Keygetter: error reading: %v", err)
				//keyChan <- "!"
				conn.Close()
				break
			}
			log.Infof("Keygetter got: %s", string(buffer[0:len]))
			keyChan <- string(buffer[0:len])
		}

	}
}
