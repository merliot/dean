package dean

import "sync"

// Socketer defines a socket interface
type Socketer interface {
	// Close the socket
	Close()
	// Send the msg on the socket
	Send(*Msg) error
	// Name of socket
	String() string
	// HasTag tests if socket has tag
	HasTag(string) bool
	// AddTag adds a tag to the socket.  A socket tag is like a VLAN ID.
	AddTag(string)
	// DelTag removes a tag from the socket.
	DelTag(string)
	// SetFlag on socket
	SetFlag(uint32)
	// TestFlag returns true if flag is set
	TestFlag(uint32) bool
}

// socket implements Socketer
type socket struct {
	name  string
	mu    sync.RWMutex
	tag   map[string]bool
	flags uint32
	bus   *Bus
}

const (
	// Socket is broadcast-ready.  If flag is not set, msgs will not be
	// broadcast on this socket.
	SocketFlagBcast uint32 = 1 << iota
)

func newSocket() *socket {
	return &socket{tag: make(map[string]bool)}
}

func (s *socket) Close() {
}

func (s *socket) Send(msg *Msg) error {
	return nil
}

func (s *socket) String() string {
	return s.name
}

func (s *socket) HasTag(string) bool {
	s.mu.RWLock()
	defer s.mu.RWUnlock()
	return s.tag
}

func (s *socket) AddTag(tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tag[tag] = true
}

func (s *socket) DelTag(tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tag, tag)
}

func (s *socket) SetFlag(flag uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flags |= flag
}

func (s *socket) TestFlag(flag uint32) bool {
	s.mu.RWLock()
	defer s.mu.RWUnlock()
	return (s.flags & flag) != 0
}
