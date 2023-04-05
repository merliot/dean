package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/lora"
)

func main() {
	lora := lora.New("gps-lora-01", "gps-lora", "gps-lora")
	server := dean.NewServer(lora)
	server.DialWebSocket("user", "passwd", "ws://10.0.0.100:8080/ws/", lora.Announce())
	server.Run()
}
