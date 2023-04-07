package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/garden/rpi"
)

func main() {
	garden := rpi.New("garden-rpi", "garden", "name")

	server := dean.NewServer(garden)

	server.BasicAuth("user", "passwd")
	server.Addr = ":8085"
	server.DialWebSocket("user", "passwd", "wss://hub.merliot.net/ws/", garden.Announce())

	go server.ListenAndServe()

	server.Run()
}
