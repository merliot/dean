package main

import (
	"log"

	"github.com/merliot/dean"
	"github.com/merliot/dean/garden"
	"github.com/merliot/dean/gps/demo"
	"github.com/merliot/dean/gps/lora"
	"github.com/merliot/dean/gps/usb"
	"github.com/merliot/dean/hub"
)

func main() {
	hub := hub.New("hub01", "hub", "hub1")

	server := dean.NewServer(hub)
	server.BasicAuth("user", "passwd")

	hub.Register("gps-demo", demo.New)
	hub.Register("gps-usb", usb.New)
	hub.Register("gps-lora", lora.New)
	hub.Register("garden", garden.New)

	log.Fatal(server.ServeTLS("hub.merliot.net"))
}
