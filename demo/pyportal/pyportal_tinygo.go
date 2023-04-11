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
	mac, err := tinynet.GetHardwareAddr()
	if err != nil {
		println("Can't get hardware MAC address")
		return
	}
	p.Mac = mac.String()
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
