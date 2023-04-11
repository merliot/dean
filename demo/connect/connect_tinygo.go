//go:build tinygo

package connect

import (
	"image/color"
	"machine"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
)

func (c *Connect) Run(i *dean.Injector) {
	var msg dean.Msg
	ticker := time.NewTicker(time.Second)

	c.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, err := tinynet.GetHardwareAddr()
	if err != nil {
		println("Can't get hardware MAC address")
		return
	}
	c.Mac = mac.String()
	c.Ip, _ = tinynet.GetIPAddr()

	for {
		changed := false

		select {
		case <- ticker.C:
		}

		if changed {
			changed = false
			c.Path = "update"
			i.Inject(msg.Marshal(c))
		}
	}
}
