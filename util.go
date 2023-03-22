package main

import (
	"net"

	log "github.com/s00500/env_logger"
)

var ip = "localhost"

func getIp() {
	if localhost {
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
			ip = ipT.String()
			break
		}
	}

}
