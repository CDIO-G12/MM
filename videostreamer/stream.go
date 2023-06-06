package videostreamer

import (
	"fmt"
	"net"
	"time"

	log "github.com/s00500/env_logger"
)

type StreamType struct {
	c       chan []byte
	running bool
}

func Init_streamer() (stream StreamType) {

	stream.c = make(chan []byte, 5)

	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:5005")
	if log.Should(err) {
		return
	}

	go func() {
		for {
			conn, err := net.DialUDP("udp", nil, raddr)
			if err != nil {
				time.Sleep(time.Second * 2)
				continue
			}
			stream.running = true
			log.Info("Streaming...")
			for {
				data := <-stream.c
				//fmt.Println(len(data))
				/*if conn == nil {
					break
				}*/
				_, err := conn.Write(data)
				//fmt.Println(i, err)
				if err != nil {
					log.Warn(err)
					break
				}
			}
			stream.running = false
		}
	}()

	return
}

func (s StreamType) Send_data(data []byte) {
	//fmt.Println(data)
	select {
	case s.c <- data: // Put 2 in the channel unless it is full
	default:
		fmt.Println("Channel full. Discarding value")
	}
}
