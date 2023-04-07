package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/garden"
)

func main() {
	garden := garden.New("garden1", "garden", "name")

	server := dean.NewServer(garden)

	server.BasicAuth("user", "passwd")
	server.Addr = ":8084"
	server.DialWebSocket("user", "passwd", "wss://hub.merliot.net/ws/", garden.Announce())

	go server.ListenAndServe()

	server.Run()
}
