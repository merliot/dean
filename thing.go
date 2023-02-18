package dean

import (
	"net/http"
)

type Thinger interface {
	Handler(*Msg)
	Serve(http.ResponseWriter, *http.Request)
	Announce() *Msg
	Run(*Injector)
}

type Dispatch struct {
	Id   string
	Path string
}

type Announce struct {
	Dispatch
	Model string
	Name  string
}

func ThingAnnounce(t Thinger) *Msg {
	var msg Msg
	var ann Announce
	ann.Path, ann.Id, ann.Model, ann.Name = "announce", "", "", ""
	//ann.Path, ann.Id, ann.Model, ann.Name = "announce", t.ID(), t.Model(), t.Name()
	msg.Marshal(&ann)
	return &msg
}
