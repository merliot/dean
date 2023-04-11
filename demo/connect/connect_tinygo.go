//go:build tinygo

package connect

import (
	"crypto/rand"
	"machine"
	"time"

	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
)

func (c *Connect) Run(i *dean.Injector) {
	var msg dean.Msg
	ticker := time.NewTicker(time.Second)

	c.CPUFreq = float64(machine.CPUFrequency()) / 1000000.0
	mac, err := tinynet.GetHardwareAddr()
	if err != nil {
		println("Can't get hardware MAC address")
		return
	}
	c.Mac = mac.String()
	c.Ip, _ = tinynet.GetIPAddr()

	for {
		changed := false

		select {
		case <- ticker.C:
		}

		if changed {
			changed = false
			c.Path = "update"
			i.Inject(msg.Marshal(c))
		}
	}
}

// TODO: remove below when RNG is working on rp2040

func init() {
	rand.Reader = &reader{}
}

type reader struct{}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	var randomByte uint32
	for i := range b {
		if i%4 == 0 {
			randomByte, err = machine.GetRNG()
			if err != nil {
				return n, err
			}
		} else {
			randomByte >>= 8
		}
		b[i] = byte(randomByte)
	}
	return len(b), nil
}
