package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/connect"
)

func main() {
	thing := connect.New("demo-connect-01", "demo-connect", "connect")
	server := dean.NewServer(thing)
	server.DialWebSocket("", "", "ws://10.0.0.100/ws/", thing.Announce())
	server.Run()
}
