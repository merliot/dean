package dean

import (
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type webSocket struct {
	socket
	conn *websocket.Conn
	ping int
}

func NewWebSocket(name string, bus *Bus) *webSocket {
	return &webSocket{
		socket: socket{name, "", bus},
	}
}

func (w *webSocket) Close() {
	w.conn.Close()
}

func (w *webSocket) Send(msg *Msg) {
	if w.conn != nil {
		println("sending:", msg.src.Name(), msg.String())
		websocket.Message.Send(w.conn, string(msg.payload))
	}
}

func (w *webSocket) Dial(user, passwd, url string, announce *Msg) {
	origin := "http://localhost/"

	// Configure the websocket
	config, err := websocket.NewConfig(url, origin)
	if err != nil {
		println("Error creating config:", err.Error())
		return
	}

	if user != "" {
		// Set the basic auth header for the request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			println("Error creating request:", err.Error())
			return
		}
		req.SetBasicAuth(user, passwd)
		config.Header = req.Header
	}

	for {
		// Dial the websocket
		conn, err := websocket.DialConfig(config)
		if err == nil {
			// Send an announcement msg
			websocket.Message.Send(conn, string(announce.payload))
			// Serve websocket until EOF
			w.serve(conn)
			// Close websocket
			conn.Close()
		} else {
			println("dial error", err.Error())
		}

		// try again in a second
		time.Sleep(time.Second)
	}
}

func (w *webSocket) servePing(conn *websocket.Conn) {
	println("PING MS", w.ping)

	for {
		var msg = &Msg{bus: w.bus, src: w}
		if err := websocket.Message.Receive(conn, &msg.payload); err != nil {
			println("disconnected", err.Error())
			w.bus.unplug(w)
			w.conn = nil
			return
		}
		w.bus.receive(msg)
	}
}

func (w *webSocket) serve(conn *websocket.Conn) {
	println("connected")

	w.conn = conn
	w.bus.plugin(w)

	if w.ping > 0 {
		w.servePing(conn)
		return
	}

	for {
		var msg = &Msg{bus: w.bus, src: w}
		if err := websocket.Message.Receive(conn, &msg.payload); err != nil {
			println("disconnected", err.Error())
			w.bus.unplug(w)
			w.conn = nil
			return
		}
		w.bus.receive(msg)
	}
}
