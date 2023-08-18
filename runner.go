package dean

import (
	"fmt"
	"sync"
)

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

func (r *Runner) DialWebSocket(user, passwd, rawURL string, announce *Msg) {
	ws := newWebSocket("websocket:"+rawURL, r.bus)
	go ws.Dial(user, passwd, rawURL, announce)
}

func (r *Runner) Run() {
	r.thinger.SetFlag(ThingFlagMetal)
	r.thinger.Run(r.injector)
}
