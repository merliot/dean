package gps

import (
	"embed"
	"fmt"
	"io/fs"
	"sync"
	"time"

	"github.com/merliot/dean"
)

//go:embed index.html
var fsys embed.FS

type update struct {
	dean.Dispatch
	Foo int
}

type Gps struct {
	*dean.Thing    `json:"-"`
	dean.Dispatch
	name string
	Foo int
	mu   sync.Mutex
}

func (g *Gps) New(user, passwd, id, name string) *Gps {
	var gps Gps
	gps.Path, gps.Id, gps.name = "gps/state", id, name
	gps.Thing = dean.NewThing(user, passwd, gps.Handler, fsys)
	return &gps
}

func (g *Gps) FileSystem() fs.FS {
	return fsys
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

func (g *Gps) Announce() *dean.Msg {
	var msg dean.Msg
	var ann dean.Announce
	ann.Path, ann.Id, ann.Model, ann.Name = "announce", g.Id, "gps", g.name
	msg.Marshal(&ann)
	return &msg
}

func (g *Gps) Run() {
	for {
		var msg dean.Msg
		var up update
		up.Path, up.Foo = "update", g.Foo+1
		g.Inject(msg.Marshal(&up))
		time.Sleep(10 * time.Second)
	}
}
