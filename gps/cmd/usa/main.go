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
	server.Dial("user", "passwd", "ws://35.185.232.122/ws/", usa.Announce())

	go server.ListenAndServe()

	server.Run()
}
