package dean

import (
	"fmt"
	"sync"
)

var defaultMaxSockets = 20

type Bus struct {
	name       string
	sockets    map[Socket]bool
	socketsMu  sync.RWMutex
	socketQ    chan bool
	handlers   map[string]func(*Msg)
	handlersMu sync.RWMutex
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
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()
	if _, ok := b.handlers[tag]; !ok {
		b.handlers[tag] = handler
		return true
	}
	return false
}

func (b *Bus) Unhandle(tag string) {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()
	delete(b.handlers, tag)
}

func (b *Bus) Name() string {
	return b.name
}

func (b *Bus) MaxSockets(maxSockets int) {
	b.socketQ = make(chan bool, maxSockets)
}

func (b *Bus) plugin(s Socket) {
	fmt.Printf("--- PLUGIN %s ---\r\n", s.Name())
	b.socketQ <- true
	b.socketsMu.Lock()
	b.sockets[s] = true
	b.socketsMu.Unlock()
	b.connect(s)
}

func (b *Bus) unplug(s Socket) {
	fmt.Printf("--- UNPLUG %s ---\r\n", s.Name())
	b.socketsMu.Lock()
	delete(b.sockets, s)
	b.socketsMu.Unlock()
	b.disconnect(s)
	<-b.socketQ
}

func (b *Bus) broadcast(msg *Msg) {
	b.socketsMu.RLock()
	defer b.socketsMu.RUnlock()
	for sock := range b.sockets {
		//println("  sock tag", sock.Tag(), "name", sock.Name())
		if msg.src != sock &&
			msg.src.Tag() == sock.Tag() &&
			sock.TestFlag(SocketFlagBcast) {
			println("broadcast:", sock.Name(), msg.String())
			sock.Send(msg)
		}
	}
}

func (b *Bus) receive(msg *Msg) {
	b.handlersMu.RLock()
	defer b.handlersMu.RUnlock()
	tag := msg.src.Tag()
	if handler, ok := b.handlers[tag]; ok {
		handler(msg)
	}
}
