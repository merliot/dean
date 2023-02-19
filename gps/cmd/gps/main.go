package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
)

func main() {
	gps := gps.New("yyyyy", "gps", "gps1")

	server := dean.NewServer(gps, nil, nil)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8080"
	server.Dial("user", "passwd", "ws://localhost:8081/ws/", gps.Announce())

	server.HandleFunc("/", gps.Serve)
	server.HandleFunc("/ws/", server.Serve)

	go server.ListenAndServe()

	server.Run()
}
