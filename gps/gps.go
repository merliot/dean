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
var hfs = http.FileServer(http.FS(fs))

type update struct {
	dean.ThingMsg
	Lat  float64
	Long float64
}

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
		ThingMsg:  dean.ThingMsg{"state"},
	}
}

func (g *Gps) saveState(msg *dean.Msg) {
	msg.Unmarshal(g)
}

func (g *Gps) getState(msg *dean.Msg) {
	msg.Marshal(g).Reply()
}

func (g *Gps) update(msg *dean.Msg) {
	var up update
	msg.Unmarshal(&up)
	g.Lat = up.Lat
	g.Long = up.Long
	msg.Broadcast()
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
	println("GPS: path:", r.URL.Path)
	switch r.URL.Path {
	case "", "/", "/index.html":
		tmpl.Execute(w, g.Vitals(r))
		return
	}
	hfs.ServeHTTP(w, r)
}

func (g *Gps) Run(i *dean.Injector) {
	select{}
}

func Run(i *dean.Injector, location func() (float64, float64)) {
	for {
		var msg dean.Msg
		var up update
		lat, long := location()
		up.Path, up.Lat, up.Long = "update", lat, long
		i.Inject(msg.Marshal(&up))
		time.Sleep(time.Minute)
	}
}
