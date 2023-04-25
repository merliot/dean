package dean

import (
	"bytes"
	"embed"
	"encoding/json"
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
	String() string
	SetFlag(uint32)
	TestFlag(uint32) bool
	Lock()
	Unlock()
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
	if t.TestFlag(ThingFlagMetal) {
		println("THINGSTORE")
		storeName := t.Model() + "-" + t.Id()
		bytes, _ := json.Marshal(t)
		os.WriteFile(storeName, bytes, 0600)
	}
}

func ThingRestore(t Thinger) {
	if t.TestFlag(ThingFlagMetal) {
		println("THINGRESTORE")
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
	flags  uint32
	mu     sync.Mutex
}

func NewThing(id, model, name string) Thing {
	return Thing{id: id, model: model, name: name}
}

const (
	ThingFlagMetal uint32 = 1 << iota
)

func (t *Thing) Subscribers() Subscribers                     { return nil }
func (t *Thing) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (t *Thing) Run(*Injector)                                { select {} }
func (t *Thing) Id() string                                   { return t.id }
func (t *Thing) Model() string                                { return t.model }
func (t *Thing) Name() string                                 { return t.name }
func (t *Thing) Lock()                                        { t.mu.Lock() }
func (t *Thing) Unlock()                                      { t.mu.Unlock() }
func (t *Thing) SetFlag(flag uint32)                          { t.flags |= flag }
func (t *Thing) TestFlag(flag uint32) bool                    { return (t.flags & flag) != 0 }
func (t *Thing) IsMetal() bool                                { return t.TestFlag(ThingFlagMetal) }

func (t *Thing) String() string {
	return "[Id: " + t.id + ", Model: " + t.model + ", Name: " + t.name + "]"
}

func (t *Thing) Announce() *Msg {
	var msg Msg
	var ann = ThingMsgAnnounce{"announce", t.id, t.model, t.name}
	msg.Marshal(&ann)
	return &msg
}

func (t *Thing) ServeFS(fs embed.FS, w http.ResponseWriter, r *http.Request) {
	scheme := "wss://"
	if r.TLS == nil {
		scheme = "ws://"
	}

	println("ServeFS:", r.URL.Path, "Id:", t.id)
	switch r.URL.Path {
	case "", "/", "/index.html":
		html, _ := fs.ReadFile("index.html")
		from := []byte("{{.WebSocket}}")
		to := []byte(scheme + r.Host + "/ws/" + t.Id() + "/")
		html = bytes.ReplaceAll(html, from, to)
		w.Write(html)
		return
	}
	http.FileServer(http.FS(fs)).ServeHTTP(w, r)
}
