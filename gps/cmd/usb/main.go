package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps/usb"
)

func main() {
	usb := usb.New("gps-usb-01", "gps-usb", "gps-usb")

	server := dean.NewServer(usb)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	server.DialWebSocket("user", "passwd", "wss://hub.merliot.net/ws/", usb.Announce())

	go server.ListenAndServe()

	server.Run()
}
