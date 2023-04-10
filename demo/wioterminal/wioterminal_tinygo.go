//go:build tinygo

package wioterminal

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	_ "github.com/merliot/dean/tinynet"
)

func (w *Wio) Run(i *dean.Injector) {
	var msg dean.Msg
	ticker := time.NewTicker(time.Second)

	for {
		changed := false

		select {
		case <- ticker.C:
			freq := machine.CPUFrequency()
			if freq != w.CPUFreq {
				w.CPUFreq = freq
				changed = true
			}
		}

		if changed {
			w.Path = "update"
			msg.Marshal(w).Broadcast()
		}
	}
}
