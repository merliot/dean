//go:build tinygo

package matrix

import (
	"machine"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"github.com/merliot/dean/lora/lorae5"
)

type loraMsg struct {
	Path   string
	Rx     string
}

type runMsg struct {
	Path string
}

func (m *Matrix) Run(i *dean.Injector) {
	var msg dean.Msg
	var loraOut = make(chan []byte)

	m.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, _ := tinynet.GetHardwareAddr()
	m.Mac = mac.String()
	m.Ip, _ = tinynet.GetIPAddr()

	m.Path = "update"
	i.Inject(msg.Marshal(m))

	relay := machine.A3
	relay.Configure(machine.PinConfig{Mode: machine.PinOutput})

	lora := lorae5.New(machine.UART1, machine.UART1_TX_PIN, machine.UART1_RX_PIN, 9600)
	lora.Init()
	go lora.RxPoll(loraOut, 2000)

	for {
		select {
		case pkt := <-loraOut:
			lmsg := loraMsg{Path: "rx", Rx: string(pkt)}
			i.Inject(msg.Marshal(&lmsg))
		case msg := <-m.runChan:
			var rmsg runMsg
			msg.Unmarshal(&rmsg)
			switch rmsg.Path {
			case "great":
				var gMsg greatMsg
				msg.Unmarshal(&gMsg)
				m.Relay = gMsg.Relay
				relay.Set(m.Relay)
				msg.Broadcast()
			case "reset":
				machine.CPUReset()
			}
		}
	}
}
