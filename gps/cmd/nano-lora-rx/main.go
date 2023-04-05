// tinygo flash -target nano-rp2040 -stack-size 4kB ~/work/dean/gps/nano-lora-rx/

package main

import (
	"encoding/json"
	"machine"
	"time"

	"github.com/merliot/dean/lora/lorae5"
)

type update struct {
	Path string
	Lat  float64
	Long float64
}

func main() {
	var up update

	time.Sleep(2 * time.Second)

	lora := lorae5.New(machine.UART0, machine.UART0_TX_PIN, machine.UART0_RX_PIN, 9600)
	lora.Init()

	for {
		pkt, err := lora.Rx(2000)
		if err == nil {
			err = json.Unmarshal(pkt, &up)
			if err == nil {
				println("GOT ONE!", up.Path, up.Lat, up.Long)
			}
		}
	}
}
