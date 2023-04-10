package main

import (
	"log"

	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/demo"
	"github.com/merliot/dean/demo/wioterminal"
)

func main() {
	demo := demo.New("demo01", "demo", "demo1").(*demo.Demo)

	server := dean.NewServer(demo)

	demo.Register("demo-wio", wioterminal.New)

	log.Fatal(server.ServeTLS("demo.merliot.net"))
}
