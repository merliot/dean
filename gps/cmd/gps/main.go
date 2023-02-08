package main

import (
	"github.com/merliot/dean/gps"
)

func main () {
	var gps *gps.Gps
	gps = gps.New("user", "passwd", "yyyyy", "gps1")
	gps.Addr = ":8080"
	go gps.ListenAndServe()
	gps.Dial("user", "passwd", "ws://localhost:8081/ws/", gps.Announce())
	gps.Run()
}
