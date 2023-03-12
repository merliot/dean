package main

import (
	//"flag"
	//"log"

	"github.com/merliot/dean"
	"github.com/merliot/dean/garden"
)

func main() {
	garden := garden.New("id", "garden", "name")

	server := dean.NewServer(garden)

	server.BasicAuth("user", "passwd")
	server.Addr = ":8082"
	//server.Dial("user", "passwd", "ws://localhost:8081/ws/", garden.Announce())

	go server.ListenAndServe()

	server.Run()

	/*
	thing.Cfg.Model = "garden"
	thing.Cfg.Name = "eden"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 6000

	flag.BoolVar(&garden.Demo, "demo", false, "Run in demo mode")
	flag.UintVar(&garden.GpioRelay, "relay", garden.GpioRelay, "Relay GPIO pin")
	flag.UintVar(&garden.GpioMeter, "meter", garden.GpioMeter, "Flow meter GPIO pin")

	flag.StringVar(&thing.Cfg.MotherHost, "rhost", "", "Remote host")
	flag.StringVar(&thing.Cfg.MotherUser, "ruser", "merle", "Remote user")
	flag.BoolVar(&thing.Cfg.IsPrime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")

	flag.Parse()

	log.Fatalln(thing.Run())
	*/
}
