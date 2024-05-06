package dean

// Injector is a one-way socket used for injecting pkts onto the bus.  The pkts
// cannot be sent back (replied) on an injector socket.
type Injector struct {
	sock socket
}

func NewInjector(name string, bus *Bus) *Injector {
	i := &Injector{socket{name, "", 0, bus}}
	bus.plugin(&i.sock)
	return i
}

// Inject a pkt onto the bus
func (i *Injector) Inject(pkt *Packet) {
	pkt.bus, pkt.src = i.sock.bus, &i.sock
	i.sock.bus.receive(pkt)
}
