package main

import (
	"log"

	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/usa"
	"github.com/merliot/dean/gps/world"
	"github.com/merliot/dean/garden"
	"github.com/merliot/dean/hub"
)

func main() {
	hub := hub.New("xxxxx", "hub", "hub1")

	server := dean.NewServer(hub)
	server.BasicAuth("user", "passwd")
	server.Addr = ":80"

	hub.Register("usa", usa.New)
	hub.Register("world", world.New)
	hub.Register("garden", garden.New)

	log.Fatal(server.ListenAndServe())
}
