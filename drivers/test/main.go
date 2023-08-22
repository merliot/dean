package main

type Linker interface {
	Send([]byte) error
}

type Tcpip struct {
	link Linker
}

func NewTcpip(link Linker) Tcpip {
	return Tcpip{link: link}
}

func (t Tcpip) Doit() {
	t.link.Send([]byte("hello"))
}

type Device struct {
	Tcpip
}

func NewDevice() *Device {
	d := &Device{}
	d.Tcpip = NewTcpip(d)
	return d
}

func (d *Device) Send(pkt []byte) error {
	println(string(pkt))
	return nil
}

func main() {
	d := NewDevice()
	d.Doit()
}
