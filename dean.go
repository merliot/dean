package dean

import (
	"encoding/json"
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
	m.payload, _ = json.Marshal(v)
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

func NewBus(name string, maxSockets int, handler func(*Msg)) *Bus {
	if handler == nil {
		handler = func(*Msg){}
	}
	return &Bus{
		name:    name,
		sockets: make(map[Socket]bool),
		socketQ: make(chan bool, maxSockets),
		handler: handler,
	}
}

func (b *Bus) Name()  string {
	return b.name
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
			println("sending:", sock.Name(), msg.String())
			sock.Send(msg)
		}
	}
	b.mu.RUnlock()
}

func (b *Bus) receive(msg *Msg) {
	b.handler(msg)
}

type injector struct {
	name string
	bus *Bus
}

func NewInjector(name string, bus *Bus) *injector {
	i := &injector{
		name: name,
		bus:  bus,
	}
	bus.plugin(i)
	return i
}

func (i *injector) Send(msg *Msg) {
	// >/dev/null
}

func (i *injector) Name() string {
	return i.name
}

func (i *injector) Inject(msg *Msg) {
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
		websocket.Message.Send(w.conn, msg.payload)
	}
}

func (w *webSocket) Name() string {
	return w.name
}

func (w *webSocket) Dial(url string, announce *Msg) {
	origin := "http://localhost/"

	for {
		conn, err := websocket.Dial(url, "", origin)
		if err != nil {
			println("dial error", err.Error())
			time.Sleep(time.Second)
			continue
		}
		websocket.Message.Send(conn, announce.payload)
		w.serve(conn)
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
