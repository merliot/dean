package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/usa"
)

func main() {
	usa := usa.New("uuusssaaa", "usa", "usa1")

	server := dean.NewServer(usa)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8082"
	server.Dial("user", "passwd", "ws://localhost:8081/ws/", usa.Announce())

	go server.ListenAndServe()

	server.Run()
}
