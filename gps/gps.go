package gps

import (
	"embed"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

type update struct {
	dean.Dispatch
	Foo int
}

type Gps struct {
	dean.Dispatch
	Foo int

	model string
	name string

	fsHandler http.Handler

	mu sync.Mutex
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	var gps = Gps{
		model: model,
		name: name,
		fsHandler: http.FileServer(http.FS(fs)),
	}
	gps.Id = id
	gps.Path = "state"
	return &gps
}

func (g *Gps) Handler(msg *dean.Msg) {
	fmt.Printf("%s\n", msg.String())

	g.mu.Lock()
	defer g.mu.Unlock()

	var dis dean.Dispatch
	msg.Unmarshal(&dis)

	switch dis.Path {
	case "get/state":
		msg.Marshal(g).Reply()
	case "update":
		var up update
		msg.Unmarshal(&up)
		g.Foo = up.Foo
		msg.Broadcast()
	}
}

func (g *Gps) Serve(w http.ResponseWriter, r *http.Request) {
	g.fsHandler.ServeHTTP(w, r)
}

func (g *Gps) Announce() *dean.Msg {
	//var msg dean.Msg
	//var ann dean.Announce
	//ann.Path, ann.Id, ann.Model, ann.Name = "announce", g.Id, g.model, g.name
	//msg.Marshal(&ann)
	//return &msg
	return dean.ThingAnnounce(g)
}

func (g *Gps) Run(i *dean.Injector) {
	for {
		var msg dean.Msg
		var up update
		up.Path, up.Id, up.Foo = "update", g.Id, g.Foo+1
		i.Inject(msg.Marshal(&up))
		time.Sleep(10 * time.Second)
	}
}
