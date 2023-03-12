//go:build demo
// +build demo

package main

import (
	"flag"
	"log"
	"os/user"

	"github.com/merliot/merle"
	"github.com/merliot/projects/garden"
	"github.com/merliot/projects/garden/demo"
)

func main() {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	garden := garden.NewGarden()
	garden.Demo = true

	thing := merle.NewThing(garden)

	thing.Cfg.Id = "garden"
	thing.Cfg.Model = "garden"
	thing.Cfg.Name = "garden01"
	thing.Cfg.PortPrivate = 7000
	thing.Cfg.MotherHost = "localhost"
	thing.Cfg.MotherUser = user.Username

	go thing.Run()

	demo := merle.NewThing(demo.NewDemo())

	demo.Cfg.Id = "garden_demo"
	demo.Cfg.Model = "garden_demo"
	demo.Cfg.Name = "garden_demo"
	demo.Cfg.PortPublic = 80
	demo.Cfg.PortPrivate = 6000

	flag.UintVar(&demo.Cfg.PortPublicTLS, "TLS", 0, "TLS port")
	flag.Parse()

	log.Fatalln(demo.Run())
}
