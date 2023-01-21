package dean

import (
	"io"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type id uint64

type Msg struct {
	Src id
	Path string
	Data []byte
}

type handler func(m Msg)

type dean struct {
	handlers map[string]handler
	bus chan(Msg)
}

func New() *dean {
	return &dean{
		handlers: make(map[string]handler),
		bus: make(chan(Msg)),
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
		websocket.JSON.Send(ws, &Msg{Src: 1, Path: "path/to/msg", Data: []byte("foo")})
		d.serve(ws)
	}
}

func (d *dean) Dial(url string) {
	go d.dial(url)
}

func (d *dean) Handle(path string, h handler) {
	d.handlers[path] = h
}

func (d *dean) receive(m Msg) {
	handler, ok := d.handlers[m.Path]
	if ok {
		handler(m)
	}
}

func (d *dean) serve(ws *websocket.Conn) {
	println("connected")
	for {
		var msg Msg
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			if err == io.EOF {
				println("disconnected")
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
