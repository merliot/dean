package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/hub"
)

var server *dean.Server

func announce(path string, thing Thinger) {
	server.Handle(path, http.StripPrefix(path, thing.Serve))
}

func main () {
	hub := hub.New("xxxxx", "hub", "hub1")

	hub.Register("gps", gps.New, announce)

	server = dean.NewServer(hub)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8081"

	server.HandleFunc("/", hub.Serve)
	server.HandleFunc("/ws/", server.Serve)

	server.ListenAndServe()
}
