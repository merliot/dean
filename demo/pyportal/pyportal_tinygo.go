//go:build tinygo

package pyportal

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	_ "github.com/merliot/dean/tinynet"
)

func (p *Pyportal) Run(i *dean.Injector) {
	var msg dean.Msg
	ticker := time.NewTicker(time.Second)

	for {
		changed := false

		select {
		case <- ticker.C:
			freq := machine.CPUFrequency()
			if freq != p.CPUFreq {
				p.CPUFreq = freq
				changed = true
			}
		}

		if changed {
			changed = false
			p.Path = "update"
			i.Inject(msg.Marshal(p))
		}
	}
}
