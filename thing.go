package dean

import (
	"net/http"

	"golang.org/x/net/websocket"
)

type Thing struct {
	name string
	http.Server
	bus      *Bus
	injector *injector
}

func NewThing(name string, maxSockets int, handler func(*Msg)) *Thing {
	bus := NewBus("thing " + name, maxSockets, handler)
	t := &Thing{
		name:     name,
		bus:      bus,
		injector: NewInjector("feed", bus),
	}
	http.HandleFunc("/ws/", t.serve)
	return t
}

func (t *Thing) Dial(url string, announce *Msg) {
	client := NewWebSocket("websocket:" + url, t.bus)
	go client.Dial(url, announce)
}

func (t *Thing) serve(w http.ResponseWriter, r *http.Request) {
	ws := NewWebSocket("websocket:" + r.Host, t.bus)
	s := websocket.Server{Handler: websocket.Handler(ws.serve)}
	s.ServeHTTP(w, r)
}

func (t *Thing) Broadcast(msg *Msg) {
	t.injector.Inject(msg)
}
