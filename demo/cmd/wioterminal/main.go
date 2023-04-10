package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/wioterminal"
)

func main() {
	wio := wioterminal.New("demo-wio-01", "demo-wio", "wioterminal")

	server := dean.NewServer(wio)
	server.DialWebSocket("", "", "wss://demo.merliot.net/ws/", wio.Announce())
	server.Run()
}
