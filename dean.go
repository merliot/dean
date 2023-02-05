package dean

import (
	"io"
	"net/http"
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

func (m *Msg) Reply() {
	m.src.Send(m)
}

func (m *Msg) Broadcast() {
	m.bus.broadcast(m)
}

type Socket interface {
	Send(*Msg)
}

type Bus struct {
	mu      sync.RWMutex
	sockets map[Socket]bool
	socketQ chan bool
	handler func(*Msg)
}

func NewBus(maxSockets int, handler func(*Msg)) *Bus {
	if handler == nil {
		handler = func(*Msg){}
	}
	return &Bus{
		sockets: make(map[Socket]bool),
		socketQ: make(chan bool, maxSockets),
		handler: handler,
	}
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
		if msg.src != sock {
			sock.Send(msg)
		}
	}
	b.mu.RUnlock()
}

func (b *Bus) receive(msg *Msg) {
	b.handler(msg)
}

type injector struct {
	bus *Bus
}

func NewInjector(bus *Bus) *injector {
	i := &injector{bus: bus}
	bus.plugin(i)
	return i
}

func (i *injector) Send(msg *Msg) {
	// >/dev/null
}

func (i *injector) Inject(msg *Msg) {
	msg.bus, msg.src = i.bus, i
	i.bus.receive(msg)
}

type webSocket struct {
	bus *Bus
	conn *websocket.Conn
}

func NewWebSocket(bus *Bus) *webSocket {
	return &webSocket{bus: bus}
}

func (w *webSocket) Send(msg *Msg) {
	if w.conn != nil {
		websocket.Message.Send(w.conn, msg.payload)
	}
}

func (w *webSocket) Dial(url string) {
	origin := "http://localhost/"

	for {
		conn, err := websocket.Dial(url, "", origin)
		if err != nil {
			println("dial error", err.Error())
			time.Sleep(time.Second)
			continue
		}
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

type webSocketServer struct {
	bus *Bus
}

func NewWebSocketServer(bus *Bus) *webSocketServer {
	return &webSocketServer{bus: bus}
}

func (s *webSocketServer) Serve(w http.ResponseWriter, r *http.Request) {
	ws := NewWebSocket(s.bus)
	server := websocket.Server{Handler: websocket.Handler(ws.serve)}
	server.ServeHTTP(w, r)
}
