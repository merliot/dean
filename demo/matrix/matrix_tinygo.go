//go:build tinygo

package matrix

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
)

func (m *Matrix) Run(i *dean.Injector) {
	var msg dean.Msg

	ticker := time.NewTicker(time.Second)

	m.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, _ := tinynet.GetHardwareAddr()
	m.Mac = mac.String()
	m.Ip, _ = tinynet.GetIPAddr()

	m.Path = "update"
	i.Inject(msg.Marshal(m))

	for {
		changed := false

		select {
		case <- ticker.C:
		}

		if changed {
			changed = false
			m.Path = "update"
			i.Inject(msg.Marshal(m))
		}
	}
}
