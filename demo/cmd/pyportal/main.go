package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/pyportal"
)

func main() {
	thing := pyportal.New("demo-pyportal-01", "demo-pyportal", "pyportal")
	server := dean.NewServer(thing)
	server.DialWebSocket("", "", "ws://demo.merliot.net/ws/1000", thing.Announce())
	server.Run()
}
