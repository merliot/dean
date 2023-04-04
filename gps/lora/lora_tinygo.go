//go:build tinygo

package lora

import (
	"encoding/json"
	"machine"

	"github.com/merliot/dean"
	"github.com/merliot/dean/lora/lorae5"
	_ "github.com/merliot/dean/tinynet"
)

func (l *Lora) Run(i *dean.Injector) {
	var msg dean.Msg

	e5 := lorae5.New(machine.UART0, machine.UART0_TX_PIN, machine.UART0_RX_PIN, 9600)
	if err := e5.Init(); err != nil {
		println(err.Error())
	}

	for {
		pkt, err := e5.Rx(2000)
		if err == nil {
			err = json.Unmarshal(pkt, l)
			if err == nil {
				println("GOT ONE!", l.Path, l.Lat, l.Long)
				i.Inject(msg.Marshal(l))
			}
		}
	}
}
