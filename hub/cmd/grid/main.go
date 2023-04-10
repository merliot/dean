package main

import (
	"log"

	"github.com/merliot/dean"
	//"github.com/merliot/dean/gps/usb"
	"github.com/merliot/dean/hub/grid"
)

func main() {
	grid := grid.New("grid01", "grid", "grid1")

	server := dean.NewServer(grid)
	//server.BasicAuth("user", "passwd")
	server.Addr = ":8080"

	//hub.Register("gps-demo", demo.New)

	log.Fatal(server.ListenAndServe())
}
