package dean

type Injector struct {
	sock socket
}

func NewInjector(name string, bus *Bus) *Injector {
	i := &Injector{socket{name, "", 0, bus}}
	bus.plugin(i)
	return i
}

func (i *Injector) Inject(msg *Msg) {
	msg.bus, msg.src = i.bus, i
	i.bus.receive(msg)
}
