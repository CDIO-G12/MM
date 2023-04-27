package utils

import (
	"net"
	"os"

	log "github.com/s00500/env_logger"
)

var IP = "localhost"

func GetIp() {
	local := os.Getenv("LOCAL")
	if localhost || local != "" {
		return
	}
	name := "Ethernet"
	if wifi {
		name = "Wi-Fi"
	}
	//fmt.Println(net.Interfaces())
	nic, err := net.InterfaceByName(name)
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := nic.Addrs()
	if err != nil { // get addresses
		return
	}
	var ipT net.IP
	for _, v := range addrs {
		if ipT = v.(*net.IPNet).IP.To4(); ipT != nil {
			IP = ipT.String()
			break
		}
	}

}

func Abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func Avg(i1, i2 int) int {
	return (i1 + i2) / 2
}
