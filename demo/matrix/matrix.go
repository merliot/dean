package matrix

import (
	"embed"
	"net"
	"net/http"

	"github.com/merliot/dean"
)

//go:embed css js images index.html
var fs embed.FS

type Matrix struct {
	dean.Thing
	dean.ThingMsg
	CPUFreq float64
	Mac     string
	Ip      net.IP
	Rx      string
	Relay   bool
	lastRx  string
	ready   bool
	runChan chan *dean.Msg
}

type greatMsg struct {
	Path  string
	Relay bool
}

func New(id, model, name string) dean.Thinger {
	println("NEW MATRIX")
	return &Matrix{
		Thing:   dean.NewThing(id, model, name),
		runChan: make(chan *dean.Msg),
	}
}

func (m *Matrix) saveState(msg *dean.Msg) {
	msg.Unmarshal(m)
}

func (m *Matrix) getState(msg *dean.Msg) {
	m.Path = "state"
	msg.Marshal(m).Reply()
}

func (m *Matrix) tinyGoIsGreat(msg *dean.Msg) {
	if m.Rx == "TinyGo" {
		m.lastRx = ""
	}
	m.lastRx += m.Rx
	relay := (m.lastRx == "TinyGoIsGreat!")
	if relay != m.Relay {
		var gMsg = greatMsg{Path: "great", Relay: relay}
		m.runChan <- msg.Marshal(&gMsg)
	}
}

func (m *Matrix) rx(msg *dean.Msg) {
	msg.Unmarshal(m).Broadcast()
	if m.IsReal() {
		m.tinyGoIsGreat(msg)
	}
}

func (m *Matrix) great(msg *dean.Msg) {
	msg.Unmarshal(m).Broadcast()
}

func (m *Matrix) reset(msg *dean.Msg) {
	msg.Unmarshal(m).Broadcast()
	if m.IsReal() {
		m.runChan <- msg
	}
}

func (m *Matrix) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     m.saveState,
		"get/state": m.getState,
		"attached":  m.getState,
		"rx":        m.rx,
		"great":     m.great,
		"reset":     m.reset,
	}
}

func (m *Matrix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.ServeFS(fs, w, r)
}
