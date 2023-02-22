package hub

import (
	"embed"
	"net/http"
	"sync"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

type generator func(id, model, name string) dean.Thinger
type callback func(dean.Socket, dean.Thinger)

type Hub struct {
	dean.Thing
	dean.ThingMsg
	fsHandler http.Handler
	gens      map[string]generator // keyed by model
	cbs       map[string]callback  // keyed by model
	mu sync.Mutex
}

func New(id, model, name string) *Hub {
	return &Hub{
		Thing:     dean.NewThing(id, model, name),
		ThingMsg:  dean.ThingMsg{"state"},
		fsHandler: http.FileServer(http.FS(fs)),
		gens:      make(map[string]generator),
		cbs:       make(map[string]callback),
	}
}

func (h *Hub) Register(model string, gen generator, cb callback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.gens[model] = gen
	h.cbs[model] = cb
}

func (h *Hub) Unregister(model string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.gens, model)
	delete(h.cbs, model)
}

func (h *Hub) getState(msg *dean.Msg) {
	msg.Marshal(h).Reply()
}

func (h *Hub) announce(msg *dean.Msg) {
	var ann dean.ThingMsgAnnounce
	msg.Unmarshal(&ann)
	if gen, ok := h.gens[ann.Model]; ok {
		thing := gen(ann.Id, ann.Model, ann.Name)
		if cb, ok := h.cbs[ann.Model]; ok {
			cb(msg.Src(), thing)
		}
	}
}

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state": h.getState,
		"announce":  h.announce,
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.fsHandler.ServeHTTP(w, r)
}

func (h *Hub) Run(i *dean.Injector) {
	select {}
}
