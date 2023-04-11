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

func (b *Base) SaveState(msg *dean.Msg) {
	msg.Unmarshal(b)
}

func (b *Base) GetState(msg *dean.Msg) {
	b.Path = "state"
	msg.Marshal(b).Reply()
}

func (b *Base) Update(msg *dean.Msg) {
	msg.Unmarshal(b).Broadcast()
}
