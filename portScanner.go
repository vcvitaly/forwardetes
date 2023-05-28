package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type portState struct {
	port int
	open state
}

type state bool

func scanPort(host string, port int) portState {
	p := portState{
		port: port,
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	scanConn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return p
	}

	err = scanConn.Close()
	if err != nil {
		log.Printf("An error while closing the connection to %s: %q", address, err)
	}
	p.open = true
	return p
}
