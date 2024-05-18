package dean

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

// webSocket wraps a websocket.Conn and implements the Socketer imterface
type webSocket struct {
	socket
	url          *url.URL
	conn         *websocket.Conn
	closing      bool
	pingPeriod   time.Duration
	pingSent     time.Time
	pongRecieved bool
	pingPkt      Packet
	pongPkt      Packet
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

	w.pingPkt = Packet{message: message{Path: "ping"}}
	w.pongPkt = Packet{message: message{Path: "pong"}}

	return w
}

func (w *webSocket) Close() {
	w.closing = true
}

func (w *webSocket) send(pkt *Packet) error {
	if w.conn == nil {
		return fmt.Errorf("Send on nil connection")
	}
	data, err := json.Marshal(pkt.message)
	if err != nil {
		return err
	}
	return websocket.Message.Send(w.conn, string(data))
}

func (w *webSocket) Send(pkt *Packet) error {
	fmt.Printf("Send  %s\r\n", pkt)
	return w.send(pkt)
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

func (w *webSocket) announced(announce *Packet) bool {

	var packet = &Packet{bus: w.bus, src: w}
	var data []byte

	// Send an announcement packet
	if err := w.Send(announce); err != nil {
		fmt.Printf("Error sending announcement: %s\r\n", err.Error())
		return false
	}

	// Any packet received is an ack of the announcement
	w.conn.SetReadDeadline(time.Now().Add(time.Second))
	err := websocket.Message.Receive(w.conn, &data)
	if err == nil {
		err = json.Unmarshal(data, &packet.message)
		if err != nil {
			fmt.Printf("Error unmarshaling packet %s: %s\r\n", w, err.Error())
			return false
		}
		w.bus.receive(packet)
		return true
	}

	// Announcement was not acked
	fmt.Printf("Announcement not ACKed %s: %s\r\n", w)
	return false
}

func (w *webSocket) Dial(user, passwd string, announce *Packet, tries int) {

	cfg, err := w.newConfig(user, passwd)
	if err != nil {
		fmt.Printf("Error configuring websocket: %s\r\n", err.Error())
		return
	}

	// tries == -1 means try forever
	for i := 0; tries < 0 || i < tries; i++ {
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
	//fmt.Printf("Connecting %s\n", w)
	w.conn = conn
	w.bus.plugin(w)
}

func (w *webSocket) disconnect() {
	//fmt.Printf("Disconnecting %s\n", w)
	w.bus.unplug(w)
	w.conn = nil
}

func (w *webSocket) serve(conn *websocket.Conn) {
	w.connect(conn)
	w.serveServer()
	w.disconnect()
}

func (w *webSocket) ping() {
	w.pongRecieved = false
	w.pingSent = time.Now()
	w.send(&w.pingPkt)
}

func (w *webSocket) serveClient() {

	var data []byte

	w.ping()

	for {
		var packet = &Packet{bus: w.bus, src: w}

		if w.closing {
			fmt.Printf("Closing %s\r\n", w)
			break
		}

		w.conn.SetReadDeadline(time.Now().Add(time.Second))
		err := websocket.Message.Receive(w.conn, &data)
		if err == nil {
			err = json.Unmarshal(data, &packet.message)
			if err != nil {
				fmt.Printf("Error unmarshaling packet, skipping %s: %s\r\n", w, err.Error())
				continue
			}
			if packet.Path == w.pongPkt.Path {
				w.pongRecieved = true
			} else {
				w.bus.receive(packet)
			}
		} else if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			// allow timeout errors
		} else {
			fmt.Printf("\r\nDisconnecting %s: %s\r\n", w, err.Error())
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

	var data []byte

	pingCheck := w.pingPeriod + (4 * time.Second)
	lastRecv := time.Now()

	for {
		var packet = &Packet{bus: w.bus, src: w}

		if w.closing {
			fmt.Printf("Closing %s\r\n", w)
			break
		}

		w.conn.SetReadDeadline(time.Now().Add(time.Second))
		err := websocket.Message.Receive(w.conn, &data)
		if err == nil {
			err = json.Unmarshal(data, &packet.message)
			if err != nil {
				fmt.Printf("Error unmarshaling packet, skipping %s: %s\r\n", w, err.Error())
				continue
			}
			lastRecv = time.Now()
			if packet.Path == w.pingPkt.Path {
				// Received ping, send pong
				err := w.send(&w.pongPkt)
				if err != nil {
					fmt.Printf("Error sending pong, disconnecting %s: %s\r\n", w, err.Error())
					break
				}
			} else {
				w.bus.receive(packet)
			}
			continue
		}

		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			if time.Now().After(lastRecv.Add(pingCheck)) {
				fmt.Printf("\r\nTimeout, disconnecting %s %s %s\r\n", w, time.Now().Sub(lastRecv).String(), err.Error())
				break
			}
			continue
		}

		fmt.Printf("\r\nDisconnecting %s: %s\r\n", w, err.Error())
		break
	}
}
