package bh1750

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/dean/amlight"
	"tinygo.org/x/drivers/bh1750"
)

type Bh1750 struct {
	*amlight.Amlight
}

func New(id, model, name string) dean.Thinger {
	println("NEW BH1750")
	return &Bh1750{
		Amlight: amlight.New(id, model, name).(*amlight.Amlight),
	}
}

func (b *Bh1750) Run(i *dean.Injector) {
	var msg dean.Msg

	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := bh1750.New(machine.I2C0)
	sensor.Configure()

	for {
		println(sensor.Illuminance())
		lux := sensor.Illuminance()
		if lux != b.Lux {
			b.Lux = lux
			b.Path = "update"
			i.Inject(msg.Marshal(b))
		}
		time.Sleep(500 * time.Millisecond)
	}
}
