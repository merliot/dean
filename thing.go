package dean

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"sync"
)

type Subscribers map[string]func(*Msg)

type ThingMaker func(id, model, name string) Thinger

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
	SetReal()
	IsReal() bool
}

type Maker interface {
	Make(id, model, name string) Thinger
}

type ThingMsg struct {
	Path string
}

type ThingMsgAnnounce struct {
	Path  string
	Id    string
	Model string
	Name  string
}

type ThingMsgConnect struct {
	Path  string
	Id    string
	Model string
	Name  string
}

type ThingMsgDisconnect struct {
	Path string
	Id   string
}

func ThingStore(t Thinger) {
	if t.IsReal() {
		storeName := t.Model() + "-" + t.Id()
		bytes, _ := json.Marshal(t)
		os.WriteFile(storeName, bytes, 0600)
	}
}

func ThingRestore(t Thinger) {
	if t.IsReal() {
		storeName := t.Model() + "-" + t.Id()
		bytes, err := os.ReadFile(storeName)
		if err == nil {
			json.Unmarshal(bytes, t)
		} else {
			ThingStore(t)
		}
	}
}

type Thing struct {
	id     string
	model  string
	name   string
	mu     sync.Mutex
	isReal bool
}

func NewThing(id, model, name string) Thing {
	return Thing{id: id, model: model, name: name}
}

func (t *Thing) Subscribers() Subscribers                     { return nil }
func (t *Thing) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (t *Thing) Run(*Injector)                                { select {} }
func (t *Thing) Id() string                                   { return t.id }
func (t *Thing) Model() string                                { return t.model }
func (t *Thing) Name() string                                 { return t.name }
func (t *Thing) Lock()                                        { t.mu.Lock() }
func (t *Thing) Unlock()                                      { t.mu.Unlock() }
func (t *Thing) SetReal()                                     { t.isReal = true }
func (t *Thing) IsReal() bool                                 { return t.isReal }

func (t *Thing) Announce() *Msg {
	var msg Msg
	var ann = ThingMsgAnnounce{"announce", t.id, t.model, t.name}
	msg.Marshal(&ann)
	return &msg
}

func (t *Thing) Vitals(r *http.Request) map[string]any {
	scheme := "wss://"
	if r.TLS == nil {
		scheme = "ws://"
	}

	return map[string]any{
		"Id":        t.id,
		"Model":     t.model,
		"Name":      t.name,
		"WebSocket": template.JSEscapeString(scheme + r.Host + "/ws/" + t.id + "/"),
	}
}
