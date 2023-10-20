package dean

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// webSocket wraps a websocket.Conn and implements the Socketer imterface
type webSocket struct {
	socket
	sync.Mutex
	conn    *websocket.Conn
	closing bool
}

func newWebSocket(name string, bus *Bus) *webSocket {
	return &webSocket{
		socket: socket{name, "", 0, bus},
	}
}

func (w *webSocket) Close() {
	w.closing = true
}

func (w *webSocket) sendRaw(msg *Msg) error {
	w.Lock()
	defer w.Unlock()
	if w.conn == nil {
		return fmt.Errorf("Send on nil connection")
	}
	return websocket.Message.Send(w.conn, msg.payload)
}

func (w *webSocket) Send(msg *Msg) error {
	w.Lock()
	defer w.Unlock()
	if w.conn == nil {
		return fmt.Errorf("Send on nil connection")
	}
	if msg.src == nil {
		println("sending:", msg.String())
	} else {
		println("sending:", msg.src.Name(), msg.String())
	}
	return websocket.Message.Send(w.conn, string(msg.payload))
}

func (w *webSocket) parsePath(path string) string {
	/* get ID from /ws/[id]/ */
	parts := strings.Split(path, "/")
	if len(parts) == 4 {
		return parts[2]
	}
	return ""
}

func (w *webSocket) newConfig(user, passwd, rawURL string) (*websocket.Config, error) {
	origin := "http://localhost/"

	// Configure the websocket
	config, err := websocket.NewConfig(rawURL, origin)
	if err != nil {
		return nil, err
	}

	if user != "" {
		// Set the basic auth header for the request
		req, err := http.NewRequest("GET", rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(user, passwd)
		config.Header = req.Header
	}

	return config, nil
}

func (w *webSocket) announced(announce *Msg) bool {

	var msg = &Msg{bus: w.bus, src: w}

	// Send an announcement msg
	if err := w.Send(announce); err != nil {
		println("error sending announcement:", err.Error())
		return false
	}

	// Any msg received is an ack of the announcement
	w.conn.SetReadDeadline(time.Now().Add(time.Second))
	err := websocket.Message.Receive(w.conn, &msg.payload)
	if err == nil {
		w.bus.receive(msg)
		return true
	}

	// Announcement was not acked
	return false
}

func (w *webSocket) Dial(user, passwd, rawURL string, announce *Msg) {

	cfg, err := w.newConfig(user, passwd, rawURL)
	if err != nil {
		println("Error configuring websocket:", err.Error())
		return
	}

	for {
		// Dial the websocket
		conn, err := websocket.DialConfig(cfg)
		if err == nil {
			w.connect(conn)
			// Send announcement
			if w.announced(announce) {
				// Serve websocket until EOF or error
				w.serveClient()
			}
			w.disconnect()
			// Close websocket
			conn.Close()
		} else {
			println("dial error", err.Error())
		}

		// try again in a second
		time.Sleep(time.Second)
	}
}

func (w *webSocket) connect(conn *websocket.Conn) {
	println("connecting")
	w.conn = conn
	w.bus.plugin(w)
}

func (w *webSocket) disconnect() {
	println("disconnecting")
	w.bus.unplug(w)
	w.conn = nil
}

var pingMsg = []byte("ping")
var pongMsg = []byte("pong")

func (w *webSocket) serve(conn *websocket.Conn) {
	w.connect(conn)
	w.serveServer()
	w.disconnect()
}

func (w *webSocket) serveClient() {

	var alive = true

	for {
		var msg = &Msg{bus: w.bus, src: w}

		if w.closing {
			println("closing")
			break
		}

		w.conn.SetReadDeadline(time.Now().Add(time.Second))
		err := websocket.Message.Receive(w.conn, &msg.payload)
		if err == nil {
			if bytes.Equal(msg.payload, pongMsg) {
				alive = true
			} else {
				w.bus.receive(msg)
			}
			continue
		}

		if netErr, ok := err.(*net.OpError); ok {
			if netErr.Timeout() {
				if !alive {
					println("no pong; disconnecting")
					break
				}
				alive = false
				err := websocket.Message.Send(w.conn, string(pingMsg))
				if err != nil {
					println("error sending ping, disconnecting", err.Error())
					break
				}
				continue
			}
		}

		println("disconnecting", err.Error())
		break
	}
}

func (w *webSocket) serveServer() {

	for {
		var msg = &Msg{bus: w.bus, src: w}

		if w.closing {
			println("closing")
			break
		}

		w.conn.SetReadDeadline(time.Now().Add(4 * time.Second))
		err := websocket.Message.Receive(w.conn, &msg.payload)
		if err == nil {
			if bytes.Equal(msg.payload, pingMsg) {
				// Received ping, send pong
				err := websocket.Message.Send(w.conn, string(pongMsg))
				if err != nil {
					println("error sending pong, disconnecting", err.Error())
					break
				}
			} else {
				w.bus.receive(msg)
			}
			continue
		}

		println("disconnecting", err.Error())
		break
	}
}
