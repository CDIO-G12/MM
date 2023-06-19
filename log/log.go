package log

import (
	"fmt"
	"net"

	u "MM/utils"

	log "github.com/s00500/env_logger"
)

type Log_type struct {
	conn net.Conn
	ch   chan string
}

func Init_log(name string, port uint16) (l Log_type) {
	address := fmt.Sprintf("%s:%d", u.IP, port)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Warn(err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Warn(err)
		return
	}
	l.conn = conn
	log.Info(name, " logging on ", address)
	l.ch = make(chan string, 100)

	go l.runner()

	return
}

func (l Log_type) runner() {
	for {
		m := <-l.ch
		if l.conn == nil {
			close(l.ch)
			return
		}
		l.conn.Write([]byte(m))
	}
}

func (l Log_type) Log(message ...any) {
	l.ch <- fmt.Sprint(message...)
}

func (l Log_type) Logf(format string, a ...any) {
	l.Log(fmt.Sprintf(format, a...))
}

func (l Log_type) Info(message ...any) {
	l.Log(message...)
}
