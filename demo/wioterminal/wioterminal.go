package wioterminal

import (
	"embed"
	"net/http"

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
