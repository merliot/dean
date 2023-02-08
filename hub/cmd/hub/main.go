package main

import (
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/hub"
)

func main () {
	var hub *hub.Hub
	var gps *gps.Gps

	hub = hub.New("user", "passwd", "xxxxx", "hub1")
	hub.Register("gps", gps)

	hub.Addr = ":8081"
	hub.ListenAndServe()
}
