package wioterminal

import (
	"embed"
	"machine"
	"net/http"
	"time"

	"github.com/merliot/dean"
)

//go:embed css js index.html
var fs embed.FS

type Wio struct {
	dean.Thing
	dean.ThingMsg
	CPUFreq uint32
	TempC int32
}

func New(id, model, name string) dean.Thinger {
	println("NEW WIOTERMINAL")
	return &Wio{
		Thing: dean.NewThing(id, model, name),
	}
}

func (w *Wio) saveState(msg *dean.Msg) {
	msg.Unmarshal(w)
}

func (w *Wio) getState(msg *dean.Msg) {
	w.Path = "state"
	msg.Marshal(w).Reply()
}

func (w *Wio) update(msg *dean.Msg) {
	msg.Unmarshal(w).Broadcast()
}

func (w *Wio) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     w.saveState,
		"get/state": w.getState,
		"attached":  w.getState,
		"update":    w.update,
	}
}

func (w *Wio) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	w.ServeFS(fs, wr, r)
}

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

