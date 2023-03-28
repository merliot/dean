package bh1750

import (
	"machine"

	"github.com/merliot/dean"
	"github.com/merliot/dean/amlight"
	"tinygo.org/x/drivers/bh1750"
)

type Bh1750 struct {
	*amlight.Amlight
	sensor machine.Device `json:"-"`
}

func New(id, model, name string) dean.Thinger {
	println("NEW BH1750")
	return &Bh1750{
		Amlight: amlight.New(id, model, name).(*amlight.Amlight),
	}
}

func (b *Bh1750) Configure() {
	machine.I2C0.Configure(machine.I2CConfig{})
        b.sensor = bh1750.New(machine.I2C0)
        b.sensor.Configure()
}

func (b *Bh1750) Illuminance() int32 {
	return 0
}
