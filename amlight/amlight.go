package amlight

import (
	"embed"
	"net/http"
	"time"

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
		Thing:     dean.NewThing(id, model, name),
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
		"attached":  a.getState,
		"update":    a.update,
	}
}

func (a *Amlight) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.ServeFS(fs, w, r)
}

func (a *Amlight) Configure() {
}

func (a *Amlight) Illuminance() int32 {
	return 0
}

func (a *Amlight) Run(i *dean.Injector) {
	var msg dean.Msg

	a.Configure()

        for {
                lux := a.Illuminance()
		if lux != a.Lux {
			a.Lux = lux
			a.Path = "update"
			i.Inject(msg.Marshal(a))
		}
                time.Sleep(500 * time.Millisecond)
        }
}
