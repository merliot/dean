package dean

import (
	"encoding/json"
	"net/http"
	"io"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type Msg struct {
	bus *Bus
	src Socket
	payload []byte
}

func NewMsg(payload []byte) *Msg {
	return &Msg{payload: payload}
}

func (m *Msg) Bytes() []byte {
	return m.payload
}

func (m *Msg) String() string {
	return string(m.payload)
}

func (m *Msg) Reply() {
	println("Reply: src", m.src.Name())
	m.src.Send(m)
}

func (m *Msg) Broadcast() {
	println("Broadcast: src", m.src.Name())
	m.bus.broadcast(m)
}

func (m *Msg) Unmarshal(v any) *Msg {
	json.Unmarshal(m.payload, v)
	return m
}

func (m *Msg) Marshal(v any) *Msg {
	var err error
	m.payload, err = json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	return m
}

type Socket interface {
	Send(*Msg)
	Name() string
}

type Bus struct {
	mu      sync.RWMutex
	name    string
	sockets map[Socket]bool
	socketQ chan bool
	handler func(*Msg)
}

var defaultMaxSockets = 10

func NewBus(name string, handler func(*Msg)) *Bus {
	if handler == nil {
		handler = func(*Msg){}
	}
	return &Bus{
		name:    name,
		sockets: make(map[Socket]bool),
		socketQ: make(chan bool, defaultMaxSockets),
		handler: handler,
	}
}

func (b *Bus) Name() string {
	return b.name
}

func (b *Bus) MaxSockets(maxSockets int) {
	b.socketQ = make(chan bool, maxSockets)
}

func (b *Bus) plugin(s Socket) {
	b.socketQ <- true
	b.mu.Lock()
	b.sockets[s] = true
	b.mu.Unlock()
}

func (b *Bus) unplug(s Socket) {
	b.mu.Lock()
	delete(b.sockets, s)
	b.mu.Unlock()
	<-b.socketQ
}

func (b *Bus) broadcast(msg *Msg) {
	b.mu.RLock()
	for sock := range b.sockets {
		println("broadcast:", sock.Name())
		if msg.src != sock {
			sock.Send(msg)
		}
	}
	b.mu.RUnlock()
}

func (b *Bus) receive(msg *Msg) {
	b.handler(msg)
}

type Injector struct {
	name string
	bus *Bus
}

func NewInjector(name string, bus *Bus) *Injector {
	i := &Injector{
		name: name,
		bus:  bus,
	}
	bus.plugin(i)
	return i
}

func (i *Injector) Send(msg *Msg) {
	// >/dev/null
}

func (i *Injector) Name() string {
	return i.name
}

func (i *Injector) Inject(msg *Msg) {
	msg.bus, msg.src = i.bus, i
	i.bus.receive(msg)
}

type webSocket struct {
	name string
	bus *Bus
	conn *websocket.Conn
}

func NewWebSocket(name string, bus *Bus) *webSocket {
	return &webSocket{
		name: name,
		bus:  bus,
	}
}

func (w *webSocket) Send(msg *Msg) {
	if w.conn != nil {
		println("sending:", msg.src.Name(), msg.String())
		websocket.Message.Send(w.conn, string(msg.payload))
	}
}

func (w *webSocket) Name() string {
	return w.name
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
		if err != nil {
			println("dial error", err.Error())
			time.Sleep(time.Second)
			continue
		}
		// Send an announcement msg
		websocket.Message.Send(conn, string(announce.payload))
		// Serve websocket until EOF
		w.serve(conn)
		// Close websocket
		conn.Close()
	}
}

func (w *webSocket) serve(conn *websocket.Conn) {
	println("connected")

	w.conn = conn
	w.bus.plugin(w)
	for {
		var msg = &Msg{bus: w.bus, src: w}
		if err := websocket.Message.Receive(conn, &msg.payload); err != nil {
			if err == io.EOF {
				println("disconnected")
				w.bus.unplug(w)
				w.conn = nil
				break
			}
			println("read error", err.Error())
		}
		w.bus.receive(msg)
	}
}
