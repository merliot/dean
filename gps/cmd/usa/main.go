package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/usa"
)

func main() {
	usa := usa.New("uuusssaaa", "usa", "usa1")

	server := dean.NewServer(usa)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	//server.Dial("user", "passwd", "ws://192.168.1.213:8081/ws/", usa.Announce())

	go server.ListenAndServe()

	server.Run()
}
