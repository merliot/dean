package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/matrix"
)

func main() {
	thing := matrix.New("demo-matrix-01", "demo-matrix", "matrix")
	server := dean.NewServer(thing)
	server.DialWebSocket("", "", "wss://demo.merliot.net/ws/", thing.Announce())
	server.Run()
}
