package gps

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

var tmpl = template.Must(template.ParseFS(fs, "index.html"))
var fserv = http.FileServer(http.FS(fs))

type update struct {
	dean.ThingMsg
	Foo int
}

type Gps struct {
	dean.Thing
	dean.ThingMsg
	Foo       int
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	return &Gps{
		Thing:     dean.NewThing(id, model, name),
		ThingMsg:  dean.ThingMsg{"state"},
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

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	println("GPS: path:", r.URL.Path)
	switch r.URL.Path {
	case "", "/", "/index.html":
		tmpl.Execute(w, g.Vitals(r))
		return
	}
	fserv.ServeHTTP(w, r)
}

func (g *Gps) Run(i *dean.Injector) {
	for {
		var msg dean.Msg
		var up update
		up.Path, up.Foo = "update", g.Foo+1
		i.Inject(msg.Marshal(&up))
		time.Sleep(120 * time.Second)
	}
}
