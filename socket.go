package dean

type Socket interface {
	Close()
	Send(*Msg)
	Name() string
	Tag() string
	SetTag(string)
	SetFlag(uint32)
	TestFlag(uint32) bool
}

type socket struct {
	name  string
	tag   string
	flags uint32
	bus   *Bus
}

const (
	SocketFlagBcast uint32 = 1 << iota
)

func (s *socket) Close() {
}

func (s *socket) Send(msg *Msg) {
	// >/dev/null
}

func (s *socket) Name() string {
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
