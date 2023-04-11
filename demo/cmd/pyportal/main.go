package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/pyportal"
)

func main() {
	py := pyportal.New("demo-pyportal-01", "demo-pyportal", "pyportal")
	server := dean.NewServer(py)
	server.DialWebSocket("", "", "wss://demo.merliot.net/ws/", py.Announce())
	server.Run()
}
