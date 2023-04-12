package matrix

import (
	"embed"
	"net"
	"net/http"

	"github.com/merliot/dean"
)

//go:embed css js index.html
var fs embed.FS

type Matrix struct {
	dean.Thing
	dean.ThingMsg
	CPUFreq float64
	Mac string
	Ip net.IP
}

func New(id, model, name string) dean.Thinger {
	println("NEW WIO")
	return &Matrix{
		Thing: dean.NewThing(id, model, name),
	}
}

func (m *Matrix) saveState(msg *dean.Msg) {
	msg.Unmarshal(m)
}

func (m *Matrix) getState(msg *dean.Msg) {
	m.Path = "state"
	msg.Marshal(m).Reply()
}

func (m *Matrix) update(msg *dean.Msg) {
	msg.Unmarshal(m).Broadcast()
}

func (m *Matrix) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     m.saveState,
		"get/state": m.getState,
		"attached":  m.getState,
		"update":    m.update,
	}
}

func (m *Matrix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.ServeFS(fs, w, r)
}
