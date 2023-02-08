package hub

import (
	"embed"
	"fmt"
	"io/fs"
	"sync"

	"github.com/merliot/dean"
)

//go:embed index.html
var fsys embed.FS

type Hub struct {
	*dean.Thing    `json:"-"`
	dean.Dispatch
	name string
	mu   sync.Mutex
}

func (h *Hub) New(user, passwd, id, name string) *Hub {
	var hub Hub
	hub.Path, hub.Id, hub.name = "hub/state", id, name
	hub.Thing = dean.NewThing(user, passwd, hub.Handler, fsys)
	return &hub
}

func (h *Hub) FileSystem() fs.FS {
	return fsys
}

func (h *Hub) Handler(msg *dean.Msg) {
	fmt.Printf("%s\n", msg.String())

	h.mu.Lock()
	defer h.mu.Unlock()

	var dis dean.Dispatch
	msg.Unmarshal(&dis)

	switch dis.Path {
	case "get/state":
		msg.Marshal(h).Reply()
	case "announce":
		var ann dean.Announce
		msg.Unmarshal(&ann)
		// TODO hook in child
	}
}

func (h *Hub) Announce() *dean.Msg {
	var msg dean.Msg
	var ann dean.Announce
	ann.Path, ann.Id, ann.Model, ann.Name = "announce", h.Id, "hub", h.name
	msg.Marshal(&ann)
	return &msg
}

func (h *Hub) Run() {
	select{}
}
