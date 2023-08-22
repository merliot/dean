package main

import (
	"machine"

	console_example "github.com/merliot/dean/drivers/examples/flash/console"
	"github.com/merliot/dean/drivers/flash"
)

func main() {
	console_example.RunFor(
		flash.NewSPI(
			&machine.SPI1,
			machine.SPI1_SDO_PIN,
			machine.SPI1_SDI_PIN,
			machine.SPI1_SCK_PIN,
			machine.SPI1_CS_PIN,
		),
	)
}
