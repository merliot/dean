package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/metro"
)

func main() {
	thing := metro.New("demo-metro-01", "demo-metro", "metro")
	server := dean.NewServer(thing)
	server.DialWebSocket("", "", "ws://demo.merliot.net/ws/1000", thing.Announce())
	server.Run()
}
