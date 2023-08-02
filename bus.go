package dean

import (
	"fmt"
	"sync"

	//sync "github.com/sasha-s/go-deadlock"
)

var defaultMaxSockets = 20

type Bus struct {
	name       string
	socketsMu  sync.RWMutex
	sockets    map[Socketer]bool
	socketQ    chan bool
	handlersMu sync.RWMutex
	handlers   map[string]func(*Msg)
	connect    func(Socketer)
	disconnect func(Socketer)
}

func NewBus(name string, connect, disconnect func(Socketer)) *Bus {
	if connect == nil {
		connect = func(Socketer) { /* don't notify */ }
	}
	if disconnect == nil {
		disconnect = func(Socketer) { /* don't notify */ }
	}
	return &Bus{
		name:       name,
		sockets:    make(map[Socketer]bool),
		socketQ:    make(chan bool, defaultMaxSockets),
		handlers:   make(map[string]func(*Msg)),
		connect:    connect,
		disconnect: disconnect,
	}
}

func (b *Bus) Handle(tag string, handler func(*Msg)) bool {
	if handler == nil {
		panic("handler is nil")
	}
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

func (b *Bus) plugin(s Socketer) {
	fmt.Printf("--- PLUGIN %s ---\r\n", s.Name())
	b.socketQ <- true
	b.socketsMu.Lock()
	b.sockets[s] = true
	b.socketsMu.Unlock()
	b.connect(s)
}

func (b *Bus) unplug(s Socketer) {
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
