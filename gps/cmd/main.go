package main

import (
	"github.com/merliot/dean/gps"
)

func main () {

	gps := gps.New("user", "passwd", "GPS", 10)

	gps.Addr = ":8080"
	go gps.ListenAndServe()

	//gps.Dial("user", "passwd", "ws://localhost:8080/ws/", gps.Announce())

	gps.Run()
}
