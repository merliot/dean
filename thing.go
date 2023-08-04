package dean

import (
	"sync"
	//sync "github.com/sasha-s/go-deadlock"
)

type Thinger interface {
	Subscribers() Subscribers
	Announce() *Msg
	Run(*Injector)
	Identity() (string, string, string)
	SetOnline(bool)
	String() string
	SetFlag(uint32)
	TestFlag(uint32) bool
}

type Subscribers map[string]func(*Msg)

type Maker interface {
	Make(id, model, name string) Thinger
}

type ThingMaker func(id, model, name string) Thinger
type Makers map[string]ThingMaker // keyed by model

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

type ThingMsgCreated struct {
	Path  string
	Id    string
	Model string
	Name  string
}

type ThingMsgDeleted struct {
	Path string
	Id   string
}

type Thing struct {
	Path   string
	Id     string
	Model  string
	Name   string
	Online bool
	flags  uint32
	mu     sync.Mutex
}

func NewThing(id, model, name string) Thing {
	if !ValidId(id) || !ValidId(model) || !ValidId(name) {
		panic("something invalid: id = \"" + id + "\", model = \"" +
			model + "\", name = \"" + name + "\"")
	}
	return Thing{Id: id, Model: model, Name: name}
}

const (
	ThingFlagMetal uint32 = 1 << iota
)

func (t *Thing) Subscribers() Subscribers           { return nil }
func (t *Thing) Run(*Injector)                      { select {} }
func (t *Thing) Identity() (string, string, string) { return t.Id, t.Model, t.Name }
func (t *Thing) Lock()                              { t.mu.Lock() }
func (t *Thing) Unlock()                            { t.mu.Unlock() }
func (t *Thing) SetOnline(online bool)              { t.Online = online }
func (t *Thing) SetFlag(flag uint32)                { t.flags |= flag }
func (t *Thing) TestFlag(flag uint32) bool          { return (t.flags & flag) != 0 }
func (t *Thing) IsMetal() bool                      { return t.TestFlag(ThingFlagMetal) }

func (t *Thing) String() string {
	return "[Id: " + t.Id + ", Model: " + t.Model + ", Name: " + t.Name + "]"
}

func (t *Thing) Announce() *Msg {
	var msg Msg
	var ann = ThingMsgAnnounce{"announce", t.Id, t.Model, t.Name}
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
