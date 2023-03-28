package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/amlight/bh1750"
)

func main() {
	light := bh1750.New("bh1750-001", "bh1750", "office")

	server := dean.NewServer(light)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	//server.Dial("user", "passwd", "ws://35.185.232.122/ws/", usa.Announce())

	//go server.ListenAndServe()

	server.Run()
}
