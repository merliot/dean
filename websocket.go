package dean

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type webSocket struct {
	socket
	sync.Mutex
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

const minPingMs = int(500) // 1/2 sec

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

func (w *webSocket) announced(announce *Msg) bool {

	for i := 0; i < 60; i++ {
		var msg = &Msg{bus: w.bus, src: w}

		// Send an announcement msg
		if err := w.Send(announce); err != nil {
			break
		}

		for {
			// Any non-ping msg received is an ack of the announcement
			w.conn.SetReadDeadline(time.Now().Add(time.Second))
			err := websocket.Message.Receive(w.conn, &msg.payload)
			if err == nil {
				if !bytes.Equal(msg.payload, pingMsg) {
					w.bus.receive(msg)
					return true
				}
				// Just a ping msg; keep reading
				continue
			} else {
				// wait a bit
				time.Sleep(time.Second)
			}
			// Timed out; send another announcement
			break
		}
	}

	// Announcement was not acked
	return false
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
			w.connect(conn)
			// Send announcement
			if w.announced(announce) {
				// Serve websocket until EOF
				w.serveConn()
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

const extraDelay = time.Second

var pingMsg = []byte("ping")

func (w *webSocket) servePing() {
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
			w.sendRaw(msg)
			pingSent = time.Now()
		}

		w.conn.SetReadDeadline(time.Now().Add(pingPeriod))
		err := websocket.Message.Receive(w.conn, &msg.payload)
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

func (w *webSocket) serveConn() {

	if w.ping > 0 {
		w.servePing()
		return
	}

	for {
		var msg = &Msg{bus: w.bus, src: w}

		if w.closing {
			println("closing")
			break
		}

		w.conn.SetReadDeadline(time.Now().Add(time.Second))
		err := websocket.Message.Receive(w.conn, &msg.payload)
		if err == nil {
			if !bytes.Equal(msg.payload, pingMsg) {
				w.bus.receive(msg)
			}
			continue
		}

		if netErr, ok := err.(*net.OpError); ok {
			if netErr.Timeout() {
				continue
			}
		}

		println("disconnecting", err.Error())
		break
	}
}

func (w *webSocket) serve(conn *websocket.Conn) {
	w.connect(conn)
	w.serveConn()
	w.disconnect()
}
