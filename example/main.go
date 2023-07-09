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
	dean.ThingMsg
	Count uint
}

func New(id, model, name string) dean.Thinger {
	return &thing{Thing: dean.NewThing(id, model, name)}
}

func (t *thing) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.FS(fs)).ServeHTTP(w, r)
}

func (t *thing) getState(msg *dean.Msg) {
	t.Path = "state"
	msg.Marshal(t).Reply()
}

func (t *thing) update(msg *dean.Msg) {
	msg.Unmarshal(t).Broadcast()
}

func (t *thing) reset(msg *dean.Msg) {
	t.Path = "update"
	t.Count = 0
	msg.Marshal(t).Reply().Broadcast()
}

func (t *thing) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state": t.getState,
		"update":    t.update,
		"reset":     t.reset,
	}
}

func (t *thing) Run(i *dean.Injector) {
	var msg dean.Msg
	for {
		t.Path = "update"
		t.Count++
		i.Inject(msg.Marshal(t))
		time.Sleep(time.Second)
	}
}

func main() {
	t := New("id", "model", "name")
	server := dean.NewServer(t)
	server.Addr = ":8080"
	go server.ListenAndServe()
	server.Run()
}
