package main

import (
	"log"

	"github.com/merliot/dean"
	"github.com/merliot/dean/demo"
	"github.com/merliot/dean/demo/pyportal"
)

func main() {
	demo := demo.New("demo01", "demo", "demo1").(*demo.Demo)

	server := dean.NewServer(demo)

	demo.Register("demo-pyportal", pyportal.New)

	log.Fatal(server.ServeTLS("demo.merliot.net"))
}
