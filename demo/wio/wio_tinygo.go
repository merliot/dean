//go:build tinygo

package connect

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
)

func (c *Connect) Run(i *dean.Injector) {
	var msg dean.Msg

	ticker := time.NewTicker(time.Second)

	c.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, _ := tinynet.GetHardwareAddr()
	c.Mac = mac.String()
	c.Ip, _ = tinynet.GetIPAddr()

	c.Path = "update"
	i.Inject(msg.Marshal(c))

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
