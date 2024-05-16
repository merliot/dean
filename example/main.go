package main

import (
	"embed"
	//"log"
	"net/http"
	"time"

	"github.com/merliot/dean"
)

//go:embed index.html
var fs embed.FS

type thing struct {
	dean.Thing
	Count uint
}

func New(id, model, name string) dean.Thinger {
	return &thing{Thing: dean.NewThing(id, model, name)}
}

func (t *thing) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.FS(fs)).ServeHTTP(w, r)
}

func (t *thing) getState(msg *dean.Packet) {
	msg.SetPath("state").Marshal(t).Reply()
}

func (t *thing) update(msg *dean.Packet) {
	msg.Unmarshal(t).Broadcast()
}

func (t *thing) reset(msg *dean.Packet) {
	t.Count = 0
	msg.SetPath("update").Marshal(t).Reply().Broadcast()
}

func (t *thing) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state": t.getState,
		"update":    t.update,
		"reset":     t.reset,
	}
}

func (t *thing) Run(i *dean.Injector) {
	var pkt dean.Packet
	for {
		t.Count++
		i.Inject(pkt.SetPath("update").Marshal(t))
		time.Sleep(time.Second)
	}
}

func main() {
	t := New("id", "model", "name")
	server := dean.NewServer(t, "", "", "8080")
	server.Run()
}
