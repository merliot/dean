package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/gps/usa"
	"github.com/merliot/dean/gps/world"
	"github.com/merliot/dean/hub"
)

func main() {
	hub := hub.New("xxxxx", "hub", "hub1")

	server := dean.NewServer(hub)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8081"

	hub.Register("gps", gps.New)
	hub.Register("usa", usa.New)
	hub.Register("world", world.New)

	server.ListenAndServe()
}
