package videostreamer

import (
	"net"
	"os/exec"
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
		exec.Command("./videostreamer/video.exe").Run()
	}()

	go func() {
	outerLoop:
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
				if data == nil {
					break outerLoop
				}
				//fmt.Println(len(data))
				_, err := conn.Write(data)
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
	case s.c <- data: // Put in channel, unless channel is full
	}
}

func (s StreamType) Close() {
	s.c <- nil
}
