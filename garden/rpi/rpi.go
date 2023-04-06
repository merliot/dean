package rpi

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/garden"
	//"gobot.io/x/gobot/drivers/gpio"
	//"gobot.io/x/gobot/platforms/raspi"
)

type Rpi struct {
	*garden.Garden
}

func New(id, model, name string) dean.Thinger {
	println("NEW GARDEN RPI")
	var r Rpi
	r.Garden = garden.New(id, model, name).(*garden.Garden)
	r.Garden.PumpOn = r.PumpOn
	r.Garden.PumpOff = r.PumpOff
	return &r
}

func (r *Rpi) Run(i *dean.Injector) {
	println("YAHOO")
	r.Garden.Run(i)
}

func (r *Rpi) PumpOn(z *garden.Zone) {
	println("PUMP ON")
}

func (r *Rpi) PumpOff(z *garden.Zone) {
	println("PUMP OFF")
}
