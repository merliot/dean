package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/wio"
)

func main() {
	thing := connect.New("demo-wio-01", "demo-wio", "wio")
	server := dean.NewServer(thing)
	server.DialWebSocket("", "", "wss://demo.merliot.net/ws/", thing.Announce())
	server.Run()
}
