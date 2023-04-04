package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/lora"
)

func main() {
	lora := lora.New("gps-lora-01", "gps-lora", "gps-lora")

	server := dean.NewServer(lora)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	server.DialWebSocket("user", "passwd", "ws://hub.merliot.net/ws/", lora.Announce())

	go server.ListenAndServe()

	server.Run()
}
