//go:build tinygo

package metro

import (
	"machine"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"github.com/merliot/dean/lora/lorae5"
)

type runMsg struct {
	Path string
}

type txMsg struct {
	Path string
	Tx   string
}

func (m *Metro) Run(i *dean.Injector) {
	var msg dean.Msg

	m.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, _ := tinynet.GetHardwareAddr()
	m.Mac = mac.String()
	m.Ip, _ = tinynet.GetIPAddr()

	m.Path = "update"
	i.Inject(msg.Marshal(m))

	lora := lorae5.New(machine.UART2, machine.UART2_TX_PIN, machine.UART2_RX_PIN, 9600)
	lora.Init()

	for {
		select {
		case msg := <-m.runChan:
			var rmsg runMsg
			msg.Unmarshal(&rmsg)
			switch rmsg.Path {
			case "tx":
				var tmsg txMsg
				msg.Unmarshal(&tmsg)
				err := lora.Tx([]byte(tmsg.Tx), 1000)
				if err != nil {
					println(err.Error())
				}
			case "reset":
				machine.CPUReset()
			}
		}
	}
}
