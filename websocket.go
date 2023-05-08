package dean

import (
	"bytes"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

type webSocket struct {
	socket
	conn    *websocket.Conn
	ping    int
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

func (w *webSocket) Send(msg *Msg) {
	if w.conn != nil {
		println("sending:", msg.src.Name(), msg.String())
		websocket.Message.Send(w.conn, string(msg.payload))
	}
}

func (w *webSocket) parsePath(path string) (id string, ping int) {
	var pingMs string

	parts := strings.Split(path, "/")

	// TODO use URL params for options like ping ms

	switch len(parts) {
	case 3:
		/* /ws/ */
		/* /ws/[ping ms] */
		pingMs = parts[2]
	case 4:
		/* /ws/[id]/ */
		/* /ws/[id]/[ping ms] */
		id = parts[2]
		pingMs = parts[3]
	default:
		return
	}

	if pingMs != "" {
		ping, _ = strconv.Atoi(pingMs)
		if ping < minPingMs {
			ping = minPingMs
		}
	}

	return
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

func (w *webSocket) Dial(user, passwd, rawURL string, announce *Msg) {

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		println("Error parsing URL:", err.Error())
		return
	}
	_, w.ping = w.parsePath(parsedURL.Path)

	cfg, err := w.newConfig(user, passwd, rawURL)
	if err != nil {
		println("Error configuring websocket:", err.Error())
		return
	}

	for {
		// Dial the websocket
		conn, err := websocket.DialConfig(cfg)
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

const extraDelay = time.Second

func (w *webSocket) servePing(conn *websocket.Conn) {
	var pingMsg = []byte{0x42, 0x42, 0x42, 0x42}
	var pingPeriod = time.Duration(w.ping) * time.Millisecond
	var quietPeriod = 2*pingPeriod + extraDelay
	var pingSent = time.Now()
	var lastRecv = pingSent

	for {
		var msg = &Msg{bus: w.bus, src: w}

		if w.closing {
			println("closing")
			return
		}

		if time.Now().Sub(pingSent) > pingPeriod {
			msg.payload = pingMsg
			websocket.Message.Send(conn, msg.payload)
			pingSent = time.Now()
		}

		conn.SetReadDeadline(time.Now().Add(pingPeriod))
		err := websocket.Message.Receive(conn, &msg.payload)
		if err == nil {
			lastRecv = time.Now()
			if !bytes.Equal(msg.payload, pingMsg) {
				w.bus.receive(msg)
			}
			continue
		}

		if netErr, ok := err.(*net.OpError); ok {
			if netErr.Timeout() {
				if time.Now().Sub(lastRecv) < quietPeriod {
					continue
				}
			}
		}

		println("disconnecting", err.Error())
		return
	}
}

func (w *webSocket) serve(conn *websocket.Conn) {
	w.connect(conn)

	if w.ping > 0 {
		w.servePing(conn)
		w.disconnect()
		return
	}

loop:
	for {
		var msg = &Msg{bus: w.bus, src: w}

		if w.closing {
			println("closing")
			break loop
		}

		conn.SetReadDeadline(time.Now().Add(time.Second))
		err := websocket.Message.Receive(conn, &msg.payload)
		if err == nil {
			w.bus.receive(msg)
			continue
		}

		if netErr, ok := err.(*net.OpError); ok {
			if netErr.Timeout() {
				continue
			}
		}

		println("disconnecting", err.Error())
		break loop
	}

	w.disconnect()
}
