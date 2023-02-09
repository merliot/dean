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
