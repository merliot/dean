package dean

import (
	"sync"
	//sync "github.com/sasha-s/go-deadlock"
)

// Thinger defines a thing interface
type Thinger interface {
	Subscribers() Subscribers
	Announce() *Msg
	Run(*Injector)
	Identity() (string, string, string)
	IsOnline() bool
	SetOnline(bool)
	String() string
	SetFlag(uint32)
	TestFlag(uint32) bool
}

type Subscribers map[string]func(*Msg)

// Maker can make a Thing
type Maker interface {
	ThingMaker
}

// ThingMaker returns a Thinger
type ThingMaker func(id, model, name string) Thinger
// Makers is a map of ThinkMakers, keyed by model
type Makers map[string]ThingMaker

// ThingMsg is the prototypical msg.  All msgs have the Path member.
type ThingMsg struct {
	Path string
}

// ThingMsgAnnounce is sent to annouce a Thing to a server
type ThingMsgAnnounce struct {
	Path  string
	Id    string
	Model string
	Name  string
}

// ThingMsgConnect is sent when a Thing connects to a server
type ThingMsgConnect struct {
	Path  string
	Id    string
	Model string
	Name  string
}

// ThingMsgDisconnect is sent when a Thing disconnects from a server
type ThingMsgDisconnect struct {
	Path string
	Id   string
}

// ThingMsgCreated is sent when as new Thing is created on a server
type ThingMsgCreated struct {
	Path  string
	Id    string
	Model string
	Name  string
}

// ThingMsgDeleted is sent when Thing is deleted from a server
type ThingMsgDeleted struct {
	Path string
	Id   string
}

// Thing implements Thinger and is the base structure for building things
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
	if !ValidId(id) || !ValidId(model) || name == "" {
		panic("something invalid: id = \"" + id + "\", model = \"" +
			model + "\", name = \"" + name + "\"")
	}
	return Thing{Id: id, Model: model, Name: name}
}

const (
	// ThingFlagMetal indicates thing is running the Run() loop
	ThingFlagMetal uint32 = 1 << iota
)

func (t *Thing) Subscribers() Subscribers           { return nil }
func (t *Thing) Run(*Injector)                      { select {} }
func (t *Thing) Identity() (string, string, string) { return t.Id, t.Model, t.Name }
func (t *Thing) Lock()                              { t.mu.Lock() }
func (t *Thing) Unlock()                            { t.mu.Unlock() }
func (t *Thing) IsOnline() bool                     { return t.Online }
func (t *Thing) SetOnline(online bool)              { t.Online = online }
func (t *Thing) SetFlag(flag uint32)                { t.flags |= flag }
func (t *Thing) TestFlag(flag uint32) bool          { return (t.flags & flag) != 0 }
func (t *Thing) IsMetal() bool                      { return t.TestFlag(ThingFlagMetal) }

func (t *Thing) String() string {
	return "[Id: " + t.Id + ", Model: " + t.Model + ", Name: " + t.Name + "]"
}

// Announce returns an announcement msg.  The announcement msg identifies the
// Thing.
func (t *Thing) Announce() *Msg {
	var msg Msg
	var ann = ThingMsgAnnounce{"announce", t.Id, t.Model, t.Name}
	return msg.Marshal(&ann)
}

// ValidId is a non-empty string with only [a-z], [A-Z], [0-9], or
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
