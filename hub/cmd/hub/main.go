package main

import (
	"github.com/merliot/dean/hub"
)

func main () {
	var hub *hub.Hub
	hub = hub.New("user", "passwd", "xxxxx", "hub1")
	hub.Addr = ":8081"
	hub.ListenAndServe()
}
