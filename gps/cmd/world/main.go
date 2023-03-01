package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/world"
)

func main() {
	world := world.New("yyyyz", "world", "world1")

	server := dean.NewServer(world)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	server.Dial("user", "passwd", "ws://localhost:8081/ws/", world.Announce())

	go server.ListenAndServe()

	server.Run()
}
