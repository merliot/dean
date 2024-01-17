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

// webSocket wraps a websocket.Conn and implements the Socketer imterface
type webSocket struct {
	socket
	sync.Mutex
	url          *url.URL
	conn         *websocket.Conn
	closing      bool
	pingPeriod   time.Duration
	pingSent     time.Time
	pongRecieved bool
}

const pingPeriodMin = time.Second

func newWebSocket(url *url.URL, remoteAddr string, bus *Bus) *webSocket {
	w := &webSocket{}

	var name string
	if remoteAddr == "" {
		name = "ws:localhost::" + url.String()
	} else {
		name = "ws:" + url.String() + "::" + remoteAddr
	}

	w.socket = socket{name, "", 0, bus}
	w.url = url

	/* param ping-period */
	period, _ := strconv.Atoi(url.Query().Get("ping-period"))
	w.pingPeriod = time.Duration(period) * time.Second
	if w.pingPeriod < pingPeriodMin {
		w.pingPeriod = pingPeriodMin
	}

	return w
}

func (w *webSocket) Close() {
	w.closing = true
}

func (w *webSocket) Send(msg *Msg) error {
	w.Lock()
	defer w.Unlock()
	if w.conn == nil {
		return fmt.Errorf("Send on nil connection")
	}
	fmt.Printf("Sending %s: %s\r\n", w, msg)
	return websocket.Message.Send(w.conn, string(msg.payload))
}

func (w *webSocket) getId() string {
	/* get ID from /ws/[id]/ */
	parts := strings.Split(w.url.Path, "/")
	if len(parts) == 4 {
		return parts[2]
	}
	return ""
}

func (w *webSocket) newConfig(user, passwd string) (*websocket.Config, error) {
	url := w.url.String()
	origin := "http://localhost/"

	// Configure the websocket
	config, err := websocket.NewConfig(url, origin)
	if err != nil {
		return nil, err
	}

	if user != "" {
		// Set the basic auth header for the request
		req, err := http.NewRequest("GET", url, nil)
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
		fmt.Printf("Error sending announcement: %s\r\n", err.Error())
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

func (w *webSocket) Dial(user, passwd string, announce *Msg) {

	cfg, err := w.newConfig(user, passwd)
	if err != nil {
		fmt.Printf("Error configuring websocket: %s\r\n", err.Error())
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
			fmt.Printf("Dial error %s: %s\r\n", w, err.Error())
		}

		// try again in a second
		time.Sleep(time.Second)
	}
}

func (w *webSocket) connect(conn *websocket.Conn) {
	fmt.Printf("Connecting %s\n", w)
	w.conn = conn
	w.bus.plugin(w)
}

func (w *webSocket) disconnect() {
	fmt.Printf("Disconnecting %s\n", w)
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

func (w *webSocket) ping() {
	w.pongRecieved = false
	w.pingSent = time.Now()
	websocket.Message.Send(w.conn, string(pingMsg))
}

func (w *webSocket) serveClient() {

	w.ping()

	for {
		var msg = &Msg{bus: w.bus, src: w}

		if w.closing {
			fmt.Printf("Closing %s\r\n", w)
			break
		}

		w.conn.SetReadDeadline(time.Now().Add(time.Second))
		err := websocket.Message.Receive(w.conn, &msg.payload)
		if err == nil {
			if bytes.Equal(msg.payload, pongMsg) {
				w.pongRecieved = true
			} else {
				w.bus.receive(msg)
			}
		} else if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			// allow timeout errors
		} else {
			fmt.Printf("Disconnecting %s: %s\r\n", w, err.Error())
			break
		}

		if time.Now().After(w.pingSent.Add(w.pingPeriod)) {
			if !w.pongRecieved {
				fmt.Printf("No pong; disconnecting %s\r\n", w)
				break
			}
			w.ping()
		}
	}
}

func (w *webSocket) serveServer() {

	pingCheck := w.pingPeriod + (4 * time.Second)
	lastRecv := time.Now()

	for {
		var msg = &Msg{bus: w.bus, src: w}

		if w.closing {
			fmt.Printf("Closing %s\r\n", w)
			break
		}

		w.conn.SetReadDeadline(time.Now().Add(time.Second))
		err := websocket.Message.Receive(w.conn, &msg.payload)
		if err == nil {
			lastRecv = time.Now()
			if bytes.Equal(msg.payload, pingMsg) {
				// Received ping, send pong
				err := websocket.Message.Send(w.conn, string(pongMsg))
				if err != nil {
					fmt.Printf("Error sending pong, disconnecting %s: %s\r\n", w, err.Error())
					break
				}
			} else {
				w.bus.receive(msg)
			}
			continue
		}

		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			if time.Now().After(lastRecv.Add(pingCheck)) {
				fmt.Printf("Timeout, disconnecting %s %s %s\r\n", w, time.Now().Sub(lastRecv).String(), err.Error())
				break
			}
			continue
		}

		fmt.Printf("Disconnecting %s: %s\r\n", w, err.Error())
		break
	}
}
