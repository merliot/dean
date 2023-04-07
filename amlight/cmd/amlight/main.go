package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/amlight"
)

func main() {
	light := amlight.New("light1", "amlight", "am21")

	server := dean.NewServer(light)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	//server.Dial("user", "passwd", "wss://hub.merliot.net/ws/", usa.Announce())

	go server.ListenAndServe()

	server.Run()
}
