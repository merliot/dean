package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/garden"
)

func main() {
	garden := garden.New("garden1", "garden", "name")

	server := dean.NewServer(garden)

	server.BasicAuth("user", "passwd")
	server.Addr = ":8082"
	server.Dial("user", "passwd", "ws://localhost:8081/ws/", garden.Announce())

	go server.ListenAndServe()

	server.Run()
}
