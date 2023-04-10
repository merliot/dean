//go:build tinygo

package wioterminal

import (
	"machine"
	"time"

	"github.com/merliot/dean"
)

func (w *Wio) Run(i *dean.Injector) {
	var msg dean.Msg
	ticker := time.NewTicker(time.Second)

	for {
		changed := false

		select {
		case <- ticker.C:
			freq = machine.CPUFrequency()
			temp = machine.ReadTempurature()
			if freq != w.CPUFreq {
				w.CPUFreq = freq
				change = true
			}
			if temp != w.TempC {
				w.TempC = temp
				changed = true
			}
		}

		if changed {
			w.Path = "update"
			msg.Marshal(w).Broadcast()
		}
	}
}
