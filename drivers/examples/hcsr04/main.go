package main

import (
	"machine"
	"time"

	"github.com/merliot/dean/drivers/hcsr04"
)

func main() {
	sensor := hcsr04.New(machine.D10, machine.D9)
	sensor.Configure()

	println("Ultrasonic starts")
	for {
		println("Distance:", sensor.ReadDistance(), "mm")

		time.Sleep(100 * time.Millisecond)
	}
}
