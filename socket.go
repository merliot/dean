package dean

// Socketer defines a socket interface
type Socketer interface {
	// Close the socket
	Close()
	// Send the pkt on the socket
	Send(*Packet) error
	// Name of socket
	String() string
	// Tag returns the socket tag
	Tag() string
	// SetTag set the socket tag.  A socket tag is like a VLAN ID.
	SetTag(string)
	// SetFlag on socket
	SetFlag(uint32)
	// TestFlag returns true if flag is set
	TestFlag(uint32) bool
}

// socket implements Socketer
type socket struct {
	name  string
	tag   string
	flags uint32
	bus   *Bus
}

const (
	// Socket is broadcast-ready.  If flag is not set, pkts will not be
	// broadcast on this socket.
	SocketFlagBcast uint32 = 1 << iota
)

func (s *socket) Close() {
}

func (s *socket) Send(pkt *Packet) error {
	return nil
}

func (s *socket) String() string {
	return s.name
}

func (s *socket) Tag() string {
	return s.tag
}

func (s *socket) SetTag(tag string) {
	s.tag = tag
}

func (s *socket) SetFlag(flag uint32) {
	s.flags |= flag
}

func (s *socket) TestFlag(flag uint32) bool {
	return (s.flags & flag) != 0
}
