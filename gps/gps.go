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
var fs embed.FS

var dispatch = struct {
	Path string
}{}

var state = struct {
	Path string
	Foo  int
}{
	Path: "state",
}

var announce = struct {
	Path  string
	Model string
	Id    string
	Name  string
}{
	Path: "announce",
	Model: "foo",
}

type update struct {
	Path string
	Foo int
}

type Gps struct {
	*dean.Thing    `json:"-"`
	mu  sync.Mutex
	Path string
	Foo int
}

func New(user, passwd, name string, maxSockets int) *Gps {
	var gps = Gps{
		Path: "state",
	}
	gps.Thing = dean.NewThing(user, passwd, name, maxSockets, gps.Handler, fs)
	return &gps
}

func (g *Gps) New() *Gps {
	return g
}

func (g *Gps) FileSystem() fs.FS {
	return fs
}

func (g *Gps) Handler(msg *dean.Msg) {
	fmt.Printf("%s\n", msg.String())

	g.mu.Lock()
	defer g.mu.Unlock()

	msg.Unmarshal(&dispatch)

	switch dispatch.Path {
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
	var ann dean.Msg
	ann.Marshal(&announce)
	return &ann
}

func (g *Gps) Run() {
	for {
		var msg dean.Msg
		var up = update{Path: "update", Foo: g.Foo+1,}
		g.Inject(msg.Marshal(&up))
		time.Sleep(10 * time.Second)
	}
}
