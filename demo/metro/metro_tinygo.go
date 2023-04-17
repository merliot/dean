//go:build tinygo

package metro

import (
	"machine"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
)

func (m *Metro) Run(i *dean.Injector) {
	var msg dean.Msg

	m.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, _ := tinynet.GetHardwareAddr()
	m.Mac = mac.String()
	m.Ip, _ = tinynet.GetIPAddr()

	m.Path = "update"
	i.Inject(msg.Marshal(m))

	for {
		select {
		case <-m.runChan:
			machine.CPUReset()
		}
	}
}
