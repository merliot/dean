package dean

import (
	"io"
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

func (d *dean) receive(m Msg) {
	println("Src", m.Src, "Path", m.Path, "Data", len(m.Data))
}

func (d *dean) Dial(url string) {
	origin := "http://localhost/"

	for {
		ws, err := websocket.Dial(url, "", origin)
		if err != nil {
			println("dial error", err.Error())
			time.Sleep(time.Second)
			continue
		}
		println("connected")
		for {
			var msg Msg
			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				if err == io.EOF {
					println("disconected")
					break
				}
				println("read error", err.Error())
			}
			d.receive(msg)
		}
	}
}

func (d *dean) Handle(path string, h handler) {
	d.handlers[path] = h
}

func (d *dean) Serve(port string) error {
	return nil
}

func (d *dean) ServeTLS(port string) error {
	return nil
}
