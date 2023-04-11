//go:build tinygo

package pyportal

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
)

func (p *Pyportal) Run(i *dean.Injector) {
	var msg dean.Msg
	ticker := time.NewTicker(time.Second)

	p.CPUFreq = machine.CPUFrequency()
	p.Mac, _ = tinynet.GetHardwareAddr()
	p.Ip, _ = tinynet.GetIPAddr()

	for {
		changed := false

		select {
		case <- ticker.C:
		}

		if changed {
			changed = false
			p.Path = "update"
			i.Inject(msg.Marshal(p))
		}
	}
}
