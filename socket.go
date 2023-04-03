package dean

type Socket interface {
	Close()
	Send(*Msg)
	Name() string
	Tag() string
	SetTag(string)
}

type socket struct {
	name string
	tag  string
	bus  *Bus
}

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
