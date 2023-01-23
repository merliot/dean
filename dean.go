package dean

import (
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type Msg struct {
	dean *dean
	src *websocket.Conn
	Path string
	Data any
}

func (m *Msg) Reply() {
	websocket.JSON.Send(m.src, m)
}

func (m *Msg) Broadcast() {
	d := m.dean
	d.RLock()
	for conn := range d.conns {
		if m.src != conn {
			websocket.JSON.Send(conn, m)
		}
	}
	d.RUnlock()
}

type handler func(m *Msg)

type dean struct {
	sync.RWMutex
	conns map[*websocket.Conn]bool
	connQ chan bool
	handlers map[string]handler
}

func New() *dean {
	return &dean{
		conns: make(map[*websocket.Conn]bool),
		connQ: make(chan bool, 10),
		handlers: make(map[string]handler),
	}
}

func (d *dean) dial(url string) {
	origin := "http://localhost/"

	for {
		ws, err := websocket.Dial(url, "", origin)
		if err != nil {
			println("dial error", err.Error())
			time.Sleep(time.Second)
			continue
		}
		d.serve(ws)
	}
}

func (d *dean) Dial(url string) {
	go d.dial(url)
}

func (d *dean) Handle(path string, h handler) {
	d.handlers[path] = h
}

func (d *dean) receive(m *Msg) {
	time.Sleep(time.Second)
	handler, ok := d.handlers[m.Path]
	if ok {
		handler(m)
	}
}

func (d *dean) plugin(ws *websocket.Conn) {
	d.connQ <- true
	d.Lock()
	d.conns[ws] = true
	d.Unlock()
}

func (d *dean) unplug(ws *websocket.Conn) {
	d.Lock()
	delete(d.conns, ws)
	d.Unlock()
	<-d.connQ
}

func (d *dean) serve(ws *websocket.Conn) {
	println("connected")
	d.plugin(ws)
	for {
		var msg = &Msg{dean: d, src: ws}
		if err := websocket.JSON.Receive(ws, msg); err != nil {
			if err == io.EOF {
				println("disconnected")
				d.unplug(ws)
				break
			}
			println("read error", err.Error())
		}
		d.receive(msg)
	}
}

func (d *dean) Serve(w http.ResponseWriter, r *http.Request) {
	s := websocket.Server{Handler: websocket.Handler(d.serve)}
	s.ServeHTTP(w, r)

}

func (d *dean) Run(run func(*Msg)) {
	var msg = Msg{dean: d}
	run(&msg)
}
