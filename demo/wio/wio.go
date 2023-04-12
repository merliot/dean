package wio

import (
	"embed"
	"net"
	"net/http"

	"github.com/merliot/dean"
)

//go:embed css js index.html
var fs embed.FS

type Wio struct {
	dean.Thing
	dean.ThingMsg
	CPUFreq float64
	Mac string
	Ip net.IP
}

func New(id, model, name string) dean.Thinger {
	println("NEW WIO")
	return &Wio{
		Thing: dean.NewThing(id, model, name),
	}
}

func (w *Wio) saveState(msg *dean.Msg) {
	msg.Unmarshal(c)
}

func (w *Wio) getState(msg *dean.Msg) {
	c.Path = "state"
	msg.Marshal(c).Reply()
}

func (w *Wio) update(msg *dean.Msg) {
	msg.Unmarshal(c).Broadcast()
}

func (w *Wio) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     c.saveState,
		"get/state": c.getState,
		"attached":  c.getState,
		"update":    c.update,
	}
}

func (w *Wio) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	c.ServeFS(fs, wr, r)
}
