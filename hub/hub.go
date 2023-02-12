package hub

import (
	"embed"
	"fmt"
	"net/http"
	"sync"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

type generator func(id, model, name string) dean.Thinger

type Hub struct {
	dean.Dispatch

	model string
	name string

	fsHandler http.Handler

	gens   map[string] generator    // keyed by model
	//things map[string] dean.Thinger // keyed by id

	mu sync.Mutex
}

func New(id, model, name string) *Hub {
	var hub = Hub{
		model: model,
		name: name,
		gens: make(map[string] generator),
		//things: make(map[string] dean.Thinger),
		fsHandler: http.FileServer(http.FS(fs)),
	}
	hub.Id = id
	hub.Path = "state"
	return &hub
}

func (h *Hub) Register(model string, gen generator, announce func(dean.Thinger)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.gens[model] = gen
}

func (h *Hub) Unregister(model string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.gens, model)
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
		gen, ok := h.gens[ann.Model]
		if ok {
			path := "/" + ann.Id + "/"
			thing := gen(ann.Id, ann.Model, ann.Name)
			//h.things[ann.Id] = thing
			h.announce(path, thing)
		}
	}
}

func (h *Hub) Serve(w http.ResponseWriter, r *http.Request) {
	h.fsHandler.ServeHTTP(w, r)
}

func (h *Hub) Announce() *dean.Msg {
	var msg dean.Msg
	var ann dean.Announce
	ann.Path, ann.Id, ann.Model, ann.Name = "announce", h.Id, h.model, h.name
	msg.Marshal(&ann)
	return &msg
}

func (h *Hub) Run(i *dean.Injector) {
	select{}
}
