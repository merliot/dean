package dean

import (
	"html/template"
	"net/http"
	"sync"
)

type Subscribers map[string]func(*Msg)

type Thinger interface {
	Subscribers() Subscribers
	ServeHTTP(http.ResponseWriter, *http.Request)
	Announce() *Msg
	Run(*Injector)
	Id() string
	Model() string
	Name() string
	Lock()
	Unlock()
}

type ThingMsg struct {
	Path string
}

type ThingMsgAnnounce struct {
	ThingMsg
	Id    string
	Model string
	Name  string
}

func ThingAnnounce(t Thinger) *Msg {
	var msg Msg
	var ann ThingMsgAnnounce
	ann.Path, ann.Id, ann.Model, ann.Name = "announce", t.Id(), t.Model(), t.Name()
	msg.Marshal(&ann)
	return &msg
}

type Thing struct {
	id    string
	model string
	name  string
	mu    sync.Mutex
}

func NewThing(id, model, name string) Thing {
	return Thing{id: id, model: model, name: name}
}

func (t *Thing) Id() string    { return t.id }
func (t *Thing) Model() string { return t.model }
func (t *Thing) Name() string  { return t.name }
func (t *Thing) Lock()         { t.mu.Lock() }
func (t *Thing) Unlock()       { t.mu.Unlock() }

func (t *Thing) Announce() *Msg {
	var msg Msg
	var ann = ThingMsgAnnounce{ThingMsg{"announce"}, t.id, t.model, t.name}
	msg.Marshal(&ann)
	return &msg
}

func (t *Thing) Vitals(r *http.Request) map[string]any {
	scheme := "wss://"
	if r.TLS == nil {
		scheme = "ws://"
	}

	return map[string]any {
		"Id":    t.id,
		"Model": t.model,
		"Name":  t.name,
		"WebSocket": template.JSEscapeString(scheme + r.Host + "/ws/" + t.id + "/"),
	}
}
