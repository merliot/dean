package gps

import (
	"embed"
	"net/http"
	"time"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

type update struct {
	dean.ThingMsg
	Id  string
	Foo int
}

type Gps struct {
	dean.Thing
	dean.ThingMsg
	Foo int
	fsHandler http.Handler
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	return &Gps{
		Thing: dean.NewThing(id, model, name),
		ThingMsg: dean.ThingMsg{"state"},
		fsHandler: http.FileServer(http.FS(fs)),
	}
}

func (g *Gps) getState(msg *dean.Msg) {
	msg.Marshal(g).Reply()
}

func (g *Gps) update(msg *dean.Msg) {
	var up update
	msg.Unmarshal(&up)
	g.Foo = up.Foo
	msg.Broadcast()
}

func (g *Gps) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state": g.getState,
		"update":    g.update,
	}
}

func (g *Gps) Serve(w http.ResponseWriter, r *http.Request) {
	g.fsHandler.ServeHTTP(w, r)
}

func (g *Gps) Run(i *dean.Injector) {
	for {
		var msg dean.Msg
		var up update
		up.Path, up.Id, up.Foo = "update", g.Id(), g.Foo+1
		i.Inject(msg.Marshal(&up))
		time.Sleep(10 * time.Second)
	}
}
