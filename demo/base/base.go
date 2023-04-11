package base

import (
	"github.com/merliot/dean"
)

type Base struct {
	dean.Thing
	dean.ThingMsg
	CPUFreq uint32
}

func New(id, model, name string) dean.Thinger {
	println("NEW BASE")
	return &Base{
		Thing: dean.NewThing(id, model, name),
	}
}

func (b *Base) saveState(msg *dean.Msg) {
	msg.Unmarshal(b)
}

func (b *Base) getState(msg *dean.Msg) {
	b.Path = "state"
	msg.Marshal(b).Reply()
}

func (b *Base) update(msg *dean.Msg) {
	msg.Unmarshal(b).Broadcast()
}

func (b *Base) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     b.saveState,
		"get/state": b.getState,
		"attached":  b.getState,
		"update":    b.update,
	}
}
