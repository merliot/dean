package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/demo"
)

func main() {
	demo := demo.New("gps-usa-01", "gps-demo", "usa1")

	server := dean.NewServer(demo)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	server.DialWebSocket("user", "passwd", "ws://hub.merliot.net/ws/", demo.Announce())

	go server.ListenAndServe()

	server.Run()
}
