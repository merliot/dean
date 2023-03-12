package hub

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/merliot/dean"
)

//go:embed css images js index.html
var fs embed.FS

var tmpl = template.Must(template.ParseFS(fs, "index.html"))
var hfs = http.FileServer(http.FS(fs))

type Child struct {
	Path   string
	Id     string
	Model  string
	Name   string
	Online bool
}

type Hub struct {
	dean.Thing
	dean.ThingMsg
	Children  map[string]Child           // keyed by id
	makers    map[string]dean.ThingMaker // keyed by model
	fsHandler http.Handler
	mu sync.Mutex
}

func New(id, model, name string) *Hub {
	return &Hub{
		Thing:     dean.NewThing(id, model, name),
		ThingMsg:  dean.ThingMsg{"state"},
		Children:  make(map[string]Child),
		makers:    make(map[string]dean.ThingMaker),
		fsHandler: http.FileServer(http.FS(fs)),
	}
}

func (h *Hub) Register(model string, maker dean.ThingMaker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.makers[model] = maker
}

func (h *Hub) Unregister(model string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.makers, model)
}

func (h *Hub) Make(id, model, name string) dean.Thinger {
	h.mu.Lock()
	defer h.mu.Unlock()
	if maker, ok := h.makers[model]; ok {
		return maker(id, model, name)
	}
	return nil
}

func (h *Hub) getState(msg *dean.Msg) {
	msg.Marshal(h).Reply()
}

func (h *Hub) connected(msg *dean.Msg) {
	fmt.Println("======== connected ==========")
	var child Child
	msg.Unmarshal(&child)
	child.Online = true
	h.Children[child.Id] = child
	msg.Marshal(&child)
	msg.Broadcast()
}

func (h *Hub) disconnected(msg *dean.Msg) {
	var child Child
	msg.Unmarshal(&child)
	child.Online = false
	h.Children[child.Id] = child
	msg.Marshal(&child)
	msg.Broadcast()
	fmt.Println("======== disconnected ==========")
}

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":    h.getState,
		"connected":    h.connected,
		"disconnected": h.disconnected,
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	println("URL:", r.URL.Path)
	switch r.URL.Path {
	case "/", "/index.html":
		tmpl.Execute(w, h.Vitals(r))
		return
	}
	hfs.ServeHTTP(w, r)
}

func (h *Hub) Run(i *dean.Injector) {
	select {}
}
