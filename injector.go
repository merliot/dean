package dean

// Injector is a one-way socket used for injecting msgs onto the bus.  The msgs
// cannot be sent back (replied) on an injector socket.
type Injector struct {
	sock socket
}

func NewInjector(name string, bus *Bus) *Injector {
	i := &Injector{socket{name, "", 0, bus}}
	bus.plugin(&i.sock)
	return i
}

// Inject a msg onto the bus
func (i *Injector) Inject(msg *Msg) {
	msg.bus, msg.src = i.sock.bus, &i.sock
	i.sock.bus.receive(msg)
}
