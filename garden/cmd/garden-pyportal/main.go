package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/garden"
)

func main() {
	garden := garden.New("garden-pyportal", "garden", "name")
	server := dean.NewServer(garden)
	server.DialWebSocket("user", "passwd", "wss://hub.merliot.net/ws/", garden.Announce())
	server.Run()
}
