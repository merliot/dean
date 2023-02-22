package main

import (
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/hub"
)

func main() {
	var server *dean.Server

	hub := hub.New("xxxxx", "hub", "hub1")
	hub.Register("gps", gps.New, server.Register)

	server = dean.NewServer(hub)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8081"

	server.ListenAndServe()
}
