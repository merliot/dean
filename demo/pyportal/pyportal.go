package pyportal

import (
	"embed"
	"image/color"
	"net"
	"net/http"

	"github.com/merliot/dean"
)

//go:embed css js index.html
var fs embed.FS

type Pyportal struct {
	dean.Thing
	dean.ThingMsg
	CPUFreq float64
	Mac string
	Ip net.IP
	Light uint16
	NeoColor color.RGBA
	neoChan chan bool
}

func New(id, model, name string) dean.Thinger {
	println("NEW PYPORTAL")
	return &Pyportal{
		Thing: dean.NewThing(id, model, name),
		NeoColor: color.RGBA{0, 0, 0, 255},
		neoChan: make(chan bool),
	}
}

func (p *Pyportal) saveState(msg *dean.Msg) {
	msg.Unmarshal(p)
}

func (p *Pyportal) getState(msg *dean.Msg) {
	p.Path = "state"
	msg.Marshal(p).Reply()
}

func (p *Pyportal) update(msg *dean.Msg) {
	msg.Unmarshal(p).Broadcast()
}

func (p *Pyportal) neo(msg *dean.Msg) {
	msg.Unmarshal(p).Broadcast()
	if p.IsReal() {
		p.neoChan <- true
	}
}

func (p *Pyportal) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     p.saveState,
		"get/state": p.getState,
		"attached":  p.getState,
		"update":    p.update,
		"neo":       p.neo,
	}
}

func (p *Pyportal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ServeFS(fs, w, r)
}
