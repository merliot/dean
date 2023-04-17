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

func (m *Matrix) poll(lora *lorae5.LoraE5, out chan []byte) {
	for {
		pkt, err := lora.Rx(2000)
		if err == nil {
			out <- pkt
		}
	}
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

	lora := lorae5.New(machine.UART1, machine.UART1_TX_PIN, machine.UART1_RX_PIN, 9600)
	lora.Init()
	go m.poll(lora, loraOut)

	for {
		select {
		case pkt := <-loraOut:
			lmsg := loraMsg{Path: "rx", Rx: string(pkt)}
			i.Inject(msg.Marshal(&lmsg))
		case <-m.runChan:
			machine.CPUReset()
		}
	}
}
