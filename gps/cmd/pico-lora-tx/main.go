// tinygo flash -target pico -stack-size 4kB ~/work/dean/gps/pico-lora-tx/

package main

import (
	"encoding/json"
	"machine"
	"time"

	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/gps/air350"
	"github.com/merliot/dean/gps/nmea"
	"github.com/merliot/dean/lora/lorae5"
)

type update struct {
	Path  string
	Lat   float64
	Long  float64
	ready bool
}

func main() {
	var out = make(chan string, 10)
	var up = update{Path: "update"}

	time.Sleep(2 * time.Second)

	lora := lorae5.New(machine.UART1, machine.UART1_TX_PIN, machine.UART1_RX_PIN, 9600)
	lora.Init()

	air350 := air350.New(machine.UART0, machine.UART0_TX_PIN, machine.UART0_RX_PIN, 9600)
	go air350.Run(out)

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case text := <-out:
			lat, long, err := nmea.ParseGLL(text)
			if err != nil {
				break
			}
			dist := int(gps.Distance(lat, long, up.Lat, up.Long) * 100.0) // cm
			if dist < 100 /* cm */ {
				break
			}
			up.Lat, up.Long, up.ready = lat, long, true
		case <-ticker.C:
			// {"Path":"update","Lat":41.629822,"Long":-72.414941}
			if !up.ready {
				break
			}
			msg, _ := json.Marshal(&up)
			println(string(msg))
			err := lora.Tx(msg, 1000)
			if err != nil {
				println(err.Error())
			}
			up.ready = false
		}
	}
}
