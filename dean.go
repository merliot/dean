package dean

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type Msg struct {
	bus     *Bus
	src     Socket
	payload []byte
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
	println("Broadcast: tag", m.src.Tag(), m.String())
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

type Bus struct {
	mu         sync.RWMutex
	name       string
	sockets    map[Socket]bool
	socketQ    chan bool
	handlers   map[string]func(*Msg)
	connect    func(Socket)
	disconnect func(Socket)
}

var defaultMaxSockets = 10

func NewBus(name string, connect, disconnect func(Socket)) *Bus {
	if connect == nil {
		connect = func(Socket) { /* don't notify */ }
	}
	if disconnect == nil {
		disconnect = func(Socket) { /* don't notify */ }
	}
	return &Bus{
		name:       name,
		sockets:    make(map[Socket]bool),
		socketQ:    make(chan bool, defaultMaxSockets),
		handlers:   make(map[string]func(*Msg)),
		connect:    connect,
		disconnect: disconnect,
	}
}

func (b *Bus) Handle(tag string, handler func(*Msg)) bool {
	if _, ok := b.handlers[tag]; !ok {
		b.handlers[tag] = handler
		return true
	}
	return false
}

func (b *Bus) Unhandle(tag string) {
	delete(b.handlers, tag)
}

func (b *Bus) Name() string {
	return b.name
}

func (b *Bus) MaxSockets(maxSockets int) {
	b.socketQ = make(chan bool, maxSockets)
}

func (b *Bus) plugin(s Socket) {
	fmt.Printf("--- PLUGIN %s ---\n", s.Name())
	b.socketQ <- true
	b.mu.Lock()
	b.sockets[s] = true
	b.mu.Unlock()
	b.connect(s)
}

func (b *Bus) unplug(s Socket) {
	fmt.Printf("--- UNPLUG %s ---\n", s.Name())
	b.mu.Lock()
	delete(b.sockets, s)
	b.mu.Unlock()
	b.disconnect(s)
	<-b.socketQ
}

func (b *Bus) broadcast(msg *Msg) {
	b.mu.RLock()
	for sock := range b.sockets {
		println("  sock tag", sock.Tag(), "name", sock.Name())
		if msg.src != sock && msg.src.Tag() == sock.Tag() {
			println("broadcast:", sock.Name(), msg.String())
			sock.Send(msg)
		}
	}
	b.mu.RUnlock()
}

func (b *Bus) receive(msg *Msg) {
	tag := msg.src.Tag()
	if handler, ok := b.handlers[tag]; ok {
		handler(msg)
	}
}

type Socket interface {
	Close()
	Send(*Msg)
	Name() string
	Tag() string
	SetTag(string)
}

type socket struct {
	name string
	tag  string
	bus  *Bus
}

func (s *socket) Close() {
}

func (s *socket) Send(msg *Msg) {
	// >/dev/null
}

func (s *socket) Name() string {
	return s.name
}

func (s *socket) Tag() string {
	return s.tag
}

func (s *socket) SetTag(tag string) {
	s.tag = tag
}

type Injector struct {
	socket
	wire chan *Msg
}

func NewInjector(name string, bus *Bus) *Injector {
	i := &Injector{
		socket: socket{name, "", bus},
		wire:   make(chan *Msg),
	}

	bus.plugin(i)

	go func() {
		for {
			select {
			case msg := <- i.wire:
				i.bus.receive(msg)
			}
		}
	}()

	return i
}

func (i *Injector) Inject(msg *Msg) {
	msg.bus, msg.src = i.bus, i
	i.wire <- msg
}

type webSocket struct {
	socket
	conn *websocket.Conn
}

func NewWebSocket(name string, bus *Bus) *webSocket {
	return &webSocket{
		socket: socket{name, "", bus},
	}
}

func (w *webSocket) Close() {
	w.conn.Close()
	w.conn = nil
	w.bus.unplug(w)
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

func (w *webSocket) serve(conn *websocket.Conn) {
	println("connected")

	w.conn = conn
	w.bus.plugin(w)
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
