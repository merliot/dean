package dean

import (
	"net/http"

	"golang.org/x/net/websocket"
)

type Thing struct {
	http.Server
	bus      *Bus
	injector *injector
}

func NewThing(maxSockets int, handler func(*Msg)) *Thing {
	bus := NewBus(maxSockets, handler)
	t := &Thing{
		bus:      bus,
		injector: NewInjector(bus),
	}
	http.HandleFunc("/ws/", t.serve)
	return t
}

func (t *Thing) Dial(url string) {
	client := NewWebSocket(t.bus)
	go client.Dial(url)
}

func (t *Thing) serve(w http.ResponseWriter, r *http.Request) {
	ws := NewWebSocket(t.bus)
	s := websocket.Server{Handler: websocket.Handler(ws.serve)}
	s.ServeHTTP(w, r)
}

func (t *Thing) Feed(msg *Msg) {
	t.injector.Inject(msg)
}
