package main

import (
	"github.com/merliot/dean/gps"
)

func main () {

	t.fsHandler = 
	http.HandleFunc("/id/", http.FileServer(http.FS(gps.FileSystem())))
	//http.HandleFunc("/ws/id/", gps.WS)

	gps := gps.New("user", "passwd", "GPS", 10)

	gps.Addr = ":8080"
	go gps.ListenAndServe()

	//gps.Dial("user", "passwd", "ws://localhost:8080/ws/", gps.Announce())

	gps.Run()
}
