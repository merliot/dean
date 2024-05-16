package dean

import (
	"fmt"
)

var defaultMaxSockets = 200

// Bus is a logical packet broadcast bus.  Packets arrive on sockets connected
// to the bus.  A received packet can be broadcast to the other sockets, or
// replied back to sender.  A socket has a tag, and the bus segregates the
// sockets by tag.  Packets arriving on a tagged socket will be broadcast only
// to other sockets with same tag.  Think of a tag as a VLAN.  The empty tag ""
// is the default tag on the bus.
type Bus struct {
	name       string
	socketsMu  rwMutex
	sockets    map[Socketer]bool
	socketQ    chan bool
	handlersMu rwMutex
	handlers   map[string]func(*Packet)
	connect    func(Socketer)
	disconnect func(Socketer)
}

// NewBus returns a new bus with connect and disconnect callbacks
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
		handlers:   make(map[string]func(*Packet)),
		connect:    connect,
		disconnect: disconnect,
	}
}

// Handle sets the packet handler for a packet tag
func (b *Bus) Handle(tag string, handler func(*Packet)) bool {
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

// Unhandle removes the packet handle for the packet tag
func (b *Bus) Unhandle(tag string) {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()
	delete(b.handlers, tag)
}

func (b *Bus) Name() string {
	return b.name
}

// MaxSockets sets the maximum number of socket connections that can be made to
// the bus.  Any socket connection attempts past the maximum will block until
// other sockets drop.
func (b *Bus) MaxSockets(maxSockets int) {
	b.socketQ = make(chan bool, maxSockets)
}

// plugin the socket to the bus
func (b *Bus) plugin(s Socketer) {
	//fmt.Printf("--- PLUGIN %s ---\r\n", s)

	// block here when socketQ is full
	//println(len(b.socketQ))
	b.socketQ <- true

	b.socketsMu.Lock()
	b.sockets[s] = true
	b.socketsMu.Unlock()

	// call connect callback
	b.connect(s)
}

// unplug the socket from the bus
func (b *Bus) unplug(s Socketer) {
	//fmt.Printf("--- UNPLUG %s ---\r\n", s)

	b.socketsMu.Lock()
	delete(b.sockets, s)
	b.socketsMu.Unlock()

	// call disconnect callback
	b.disconnect(s)

	// release one from the socketQ
	<-b.socketQ
}

// broadcast packet to all sockets with matching tag, skipping the source
// socket src
func (b *Bus) broadcast(pkt *Packet) {
	b.socketsMu.RLock()
	defer b.socketsMu.RUnlock()
	for sock := range b.sockets {
		//println("  sock tag", sock.Tag(), "name", sock.Name())
		if pkt.src != sock &&
			pkt.src.Tag() == sock.Tag() &&
			sock.TestFlag(SocketFlagBcast) {
			fmt.Printf("Bcast  src %s dst %s packet %s\r\n", pkt.src, sock, pkt)
			sock.Send(pkt)
		}
	}
}

// receive will call the packet handler for the packet tag
func (b *Bus) receive(pkt *Packet) {
	fmt.Printf("Recv  %s\r\n", pkt)
	tag := pkt.popTag()
	b.handlersMu.RLock()
	defer b.handlersMu.RUnlock()
	if handler, ok := b.handlers[tag]; ok {
		handler(pkt)
	}
}
