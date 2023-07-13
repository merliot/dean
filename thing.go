package dean

import (
	"sync"
	//sync "github.com/sasha-s/go-deadlock"
)

type Subscribers map[string]func(*Msg)

type ThingMaker func(id, model, name string) Thinger

type Thinger interface {
	Subscribers() Subscribers
	Announce() *Msg
	Run(*Injector)
	Id() string
	Model() string
	Name() string
	String() string
	SetFlag(uint32)
	TestFlag(uint32) bool
}

type Maker interface {
	Make(id, model, name string) Thinger
}

type ThingMsg struct {
	Path string
}

type ThingMsgAnnounce struct {
	Path  string
	Id    string
	Model string
	Name  string
}

type ThingMsgConnect struct {
	Path  string
	Id    string
	Model string
	Name  string
}

type ThingMsgDisconnect struct {
	Path string
	Id   string
}

type Thing struct {
	id    string
	model string
	name  string
	flags uint32
	mu    sync.Mutex
}

func NewThing(id, model, name string) Thing {
	if !ValidId(id) || !ValidId(model) || !ValidId(name) {
		panic("something invalid: id = \"" + id + "\", model = \"" +
			model + "\", name = \"" + name + "\"")
	}
	return Thing{id: id, model: model, name: name}
}

const (
	ThingFlagMetal uint32 = 1 << iota
)

func (t *Thing) Subscribers() Subscribers                     { return nil }
func (t *Thing) Run(*Injector)                                { select {} }
func (t *Thing) Id() string                                   { return t.id }
func (t *Thing) Model() string                                { return t.model }
func (t *Thing) Name() string                                 { return t.name }
func (t *Thing) Lock()                                        { t.mu.Lock() }
func (t *Thing) Unlock()                                      { t.mu.Unlock() }
func (t *Thing) SetFlag(flag uint32)                          { t.flags |= flag }
func (t *Thing) TestFlag(flag uint32) bool                    { return (t.flags & flag) != 0 }
func (t *Thing) IsMetal() bool                                { return t.TestFlag(ThingFlagMetal) }

func (t *Thing) String() string {
	return "[Id: " + t.id + ", Model: " + t.model + ", Name: " + t.name + "]"
}

func (t *Thing) Announce() *Msg {
	var msg Msg
	var ann = ThingMsgAnnounce{"announce", t.id, t.model, t.name}
	return msg.Marshal(&ann)
}

// A valid ID is a non-empty string with only [a-z], [A-Z], [0-9], or
// underscore characters.
func ValidId(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') &&
			(r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') &&
			(r != '_') {
			return false
		}
	}
	return len(s) > 0
}
