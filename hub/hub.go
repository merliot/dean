package hub

import (
	"embed"
	"net/http"
	"sync"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

type Factory struct {
	Gen func(id, model, name string) dean.Thinger
	Ann func(dean.Socket, dean.Thinger)
}

type Hub struct {
	dean.Thing
	dean.ThingMsg
	fsHandler http.Handler
	factories map[string]Factory // keyed by model
	//things map[string] dean.Thinger // keyed by id
	mu sync.Mutex
}

func New(id, model, name string) *Hub {
	return &Hub{
		Thing: dean.NewThing(id, model, name),
		ThingMsg: dean.ThingMsg{"state"},
		//things: make(map[string] dean.Thinger),
		fsHandler: http.FileServer(http.FS(fs)),
		factories: make(map[string]Factory),
	}
}

func (h *Hub) Register(model string, factory Factory) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.factories[model] = factory
}

func (h *Hub) Unregister(model string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.factories, model)
}

func (h *Hub) getState(msg *dean.Msg) {
	msg.Marshal(h).Reply()
}

func (h *Hub) announce(msg *dean.Msg) {
	var ann dean.ThingMsgAnnounce
	msg.Unmarshal(&ann)
	factory, ok := h.factories[ann.Model]
	if ok {
		thing := factory.Gen(ann.Id, ann.Model, ann.Name)
		//h.things[ann.Id] = thing
		factory.Ann(msg.Src(), thing)
	}
}

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state": h.getState,
		"announce":  h.announce,
	}
}

func (h *Hub) Serve(w http.ResponseWriter, r *http.Request) {
	h.fsHandler.ServeHTTP(w, r)
}

func (h *Hub) Run(i *dean.Injector) {
	select {}
}
