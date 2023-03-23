package gps

import (
	"embed"
	"net/http"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

type Gps struct {
	dean.Thing
	dean.ThingMsg
	Lat  float64
	Long float64
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	return &Gps{
		Thing:     dean.NewThing(id, model, name),
	}
}

func (g *Gps) saveState(msg *dean.Msg) {
	msg.Unmarshal(g)
}

func (g *Gps) getState(msg *dean.Msg) {
	g.Path = "state"
	msg.Marshal(g).Reply()
}

func (g *Gps) update(msg *dean.Msg) {
	msg.Unmarshal(g).Broadcast()
}

func (g *Gps) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     g.saveState,
		"get/state": g.getState,
		"attached":  g.getState,
		"update":    g.update,
	}
}

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.ServeFS(fs, w, r)
}

func (g *Gps) Run(i *dean.Injector) {
	select{}
}
