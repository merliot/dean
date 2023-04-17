//go:build tinygo

package matrix

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/gps/air350"
	"github.com/merliot/dean/gps/nmea"
)

type locMsg struct {
	Path string
	Lat  float64
	Long float64
}

func (m *Matrix) Run(i *dean.Injector) {
	var airOut = make(chan string, 10)
	var msg dean.Msg

	m.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, _ := tinynet.GetHardwareAddr()
	m.Mac = mac.String()
	m.Ip, _ = tinynet.GetIPAddr()

	air350 := air350.New(machine.UART1, machine.UART1_TX_PIN, machine.UART1_RX_PIN, 9600)
	go air350.Run(airOut)

	m.Path = "update"
	i.Inject(msg.Marshal(m))

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case text := <-airOut:
			lat, long, err := nmea.ParseGLL(text)
			if err != nil {
				break
			}
			dist := int(gps.Distance(lat, long, m.Lat, m.Long) * 100.0) // cm
			if dist < 100 /* cm */ {
				break
			}
			m.Lat, m.Long, m.ready = lat, long, true
		case <-ticker.C:
			if !m.ready {
				break
			}
			// {"Path":"loc","Lat":41.629822,"Long":-72.414941}
			lmsg := locMsg{Path: "loc", Lat: m.Lat, Long: m.Long}
			i.Inject(msg.Marshal(&lmsg))
			m.ready = false
		case <-m.runChan:
			machine.CPUReset()
		}
	}
}
