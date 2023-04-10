package main

import (
	"log"

	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/grid"
	"github.com/merliot/dean/demo/wioterminal"
)

func main() {
	grid := grid.New("grid01", "grid", "grid1").(*grid.Grid)

	server := dean.NewServer(grid)
	server.Addr = ":8080"

	grid.Register("demo-wio", wioterminal.New)

	log.Fatal(server.ListenAndServe())
}
