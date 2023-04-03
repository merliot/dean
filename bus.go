package dean

import (
	"fmt"
	"sync"
)

var defaultMaxSockets = 10

type Bus struct {
	mu         sync.RWMutex
	name       string
	sockets    map[Socket]bool
	socketQ    chan bool
	handlers   map[string]func(*Msg)
	connect    func(Socket)
	disconnect func(Socket)
}

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
