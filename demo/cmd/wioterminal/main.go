package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/wioterminal"
)

func main() {
	wio := wioterminal.New("demo-wio-01", "demo-wio", "wioterminal")

	server := dean.NewServer(wio)
	server.DialWebSocket("", "", "ws://localhost:8080/ws/", wio.Announce())
	server.Run()
}
