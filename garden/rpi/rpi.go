package rpi

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/garden"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

type Rpi struct {
	*garden.Garden
	relays [4]*gpio.RelayDriver
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
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	r.relays[0] = gpio.NewRelayDriver(adaptor, "31")
	r.relays[1] = gpio.NewRelayDriver(adaptor, "33")
	r.relays[2] = gpio.NewRelayDriver(adaptor, "35")
	r.relays[3] = gpio.NewRelayDriver(adaptor, "37")

	for _, relay := range r.relays {
		relay.Start()
		relay.Off()
	}

	r.Garden.Run(i)
}

func (r *Rpi) PumpOn(z *garden.Zone) {
	r.relays[z.Index].On()
}

func (r *Rpi) PumpOff(z *garden.Zone) {
	r.relays[z.Index].Off()
}
