//go:build tinygo

package pyportal

import (
	"image/color"
	"machine"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"tinygo.org/x/drivers/ws2812"
)

func (p *Pyportal) Run(i *dean.Injector) {
	var msg dean.Msg
	ticker := time.NewTicker(time.Second)

	p.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, err := tinynet.GetHardwareAddr()
	if err != nil {
		println("Can't get hardware MAC address")
		return
	}
	p.Mac = mac.String()
	p.Ip, _ = tinynet.GetIPAddr()

	lightSensor := machine.ADC{machine.A2}
	lightSensor.Configure(machine.ADCConfig{})

	neo := machine.WS2812
	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ws := ws2812.New(neo)
	ws.WriteColors([]color.RGBA{p.NeoColor})

	for {
		changed := false

		select {
		case <- ticker.C:
			val := lightSensor.Get()
			if val != p.Light {
				p.Light = val
				changed = true
			}
		}

		if changed {
			changed = false
			p.Path = "update"
			i.Inject(msg.Marshal(p))
		}
	}
}
