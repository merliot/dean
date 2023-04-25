package amlight

import (
	"embed"
	"net/http"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

type Amlight struct {
	dean.Thing
	dean.ThingMsg
	Lux int32 // mlx (milliLux)
}

func New(id, model, name string) dean.Thinger {
	println("NEW AMLIGHT")
	return &Amlight{
		Thing: dean.NewThing(id, model, name),
	}
}

func (a *Amlight) saveState(msg *dean.Msg) {
	msg.Unmarshal(a)
}

func (a *Amlight) getState(msg *dean.Msg) {
	a.Path = "state"
	msg.Marshal(a).Reply()
}

func (a *Amlight) update(msg *dean.Msg) {
	msg.Unmarshal(a).Broadcast()
}

func (a *Amlight) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     a.saveState,
		"get/state": a.getState,
		"update":    a.update,
	}
}

func (a *Amlight) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.ServeFS(fs, w, r)
}

func (a *Amlight) Run(i *dean.Injector) {
	select {}
}
