package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/lora"
)

func main() {
	lora := lora.New("gps-lora-01", "gps-lora", "gps-lora")
	server := dean.NewServer(lora)
	server.DialWebSocket("user", "passwd", "ws://hub.merliot.net/ws/", lora.Announce())
	server.Run()
}
