package main

import (
	"net/http"

	"github.com/merliot/dean/gps"
)

func main () {
	var gps    *gps.Gps
	var server *dean.Server

	gps = gps.New("yyyyy", "gps", "gps1")
	server = dean.NewServer(gps.Handler)

	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	server.Dial("user", "passwd", "ws://localhost:8081/ws/", gps.Announce())

	server.HandleFunc("/", gps.Serve)
	server.HandleFunc("/ws/", server.Serve)

	go server.ListenAndServe()

	gps.Run()
}
