package dean

import (
	"fmt"
	"net/url"
	"sync"
)

// Runner runs a Thing
type Runner struct {
	thinger  Thinger
	bus      *Bus
	injector *Injector
}

func NewRunner(thinger Thinger) *Runner {
	var r Runner

	r.thinger = thinger
	r.thinger.SetOnline(true)

	r.bus = NewBus("runner bus", nil, nil)
	r.bus.Handle("", r.busHandle(thinger))
	r.injector = NewInjector("runner injector", r.bus)

	return &r
}

func (r *Runner) busHandle(thinger Thinger) func(*Msg) {
	return func(msg *Msg) {
		fmt.Printf("Bus handle %s\r\n", msg.String())
		var rmsg ThingMsg

		msg.Unmarshal(&rmsg)

		switch rmsg.Path {
		case "get/state", "state":
			msg.src.SetFlag(SocketFlagBcast)
		}

		if locker, ok := thinger.(sync.Locker); ok {
			locker.Lock()
			defer locker.Unlock()
		}

		subs := thinger.Subscribers()
		if sub, ok := subs[rmsg.Path]; ok {
			sub(msg)
		}
	}
}

// DialWebSocket connects a web socket to rawURL.  User/passwd can be set for HTTP
// Basic Auth.  The announce msg is sent when the web socket connects.
func (r *Runner) DialWebSocket(user, passwd, rawUrl string, announce *Msg) {
	u, _ := url.Parse(rawUrl)
	ws := newWebSocket(u, r.bus)
	go ws.Dial(user, passwd, rawUrl, announce)
}

// Run the main run loop
func (r *Runner) Run() {
	r.thinger.SetFlag(ThingFlagMetal)
	r.thinger.Run(r.injector)
}
