package main

import (
	"log"

	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/grid"
	"github.com/merliot/dean/demo/wioterminal"
)

func main() {
	demo := grid.New("grid01", "grid", "grid1").(*grid.Grid)

	server := dean.NewServer(demo)

	demo.Register("demo-wio", wioterminal.New)

	log.Fatal(server.ServeTLS("demo.merliot.net"))
}
