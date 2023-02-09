package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/hub"
)

func main () {
	hub := hub.New("xxxxx", "hub", "hub1")
	hub.Register("gps", gps.New)

	server := dean.NewServer(hub)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8081"

	server.HandleFunc("/", hub.Serve)
	server.HandleFunc("/ws/", server.Serve)

	server.HandleFunc("/yyyyy/", http.StripPrefix("/yyyyy/", gps.Serve))

	server.ListenAndServe()
}
