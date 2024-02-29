package dean

import (
	"fmt"
	"net/url"
	"strings"
)

// Runner runs a Thing
type Runner struct {
	thinger  Thinger
	bus      *Bus
	injector *Injector

	user   string
	passwd string
}

func NewRunner(thinger Thinger, user, passwd string) *Runner {
	var r Runner

	r.user = user
	r.passwd = passwd

	r.thinger = thinger
	r.thinger.SetOnline(true)

	r.bus = NewBus("runner bus", nil, nil)
	r.bus.Handle("", r.busHandle(thinger))
	r.injector = NewInjector("runner injector", r.bus)

	return &r
}

func (r *Runner) busHandle(thinger Thinger) func(*Msg) {
	return func(msg *Msg) {
		fmt.Printf("Bus handle src %s msg %s\r\n", msg.src, msg)
		var rmsg ThingMsg

		msg.Unmarshal(&rmsg)

		switch rmsg.Path {
		case "get/state", "state":
			msg.src.SetFlag(SocketFlagBcast)
		}

		if locker, ok := thinger.(Locker); ok {
			locker.Lock()
			defer locker.Unlock()
		}

		subs := thinger.Subscribers()
		if sub, ok := subs[rmsg.Path]; ok {
			sub(msg)
		}
	}
}

// Dial connects runner to other servers
func (r *Runner) Dial(dialURLs string) {
	for _, u := range strings.Split(dialURLs, ",") {
		purl, err := url.Parse(u)
		if err != nil {
			fmt.Printf("Error parsing URL: %s\r\n", err.Error())
			continue
		}
		switch purl.Scheme {
		case "ws", "wss":
			ws := newWebSocket(purl, "", r.bus)
			go ws.Dial(r.user, r.passwd, r.thinger.Announce())
		default:
			println("Scheme must be ws or wss:", u)
		}
	}
}

// Run the main run loop
func (r *Runner) Run() {
	// Thinger is metal when run in runner
	r.thinger.SetFlag(ThingFlagMetal)
	// Setup thinger
	r.thinger.Setup()
	// Run thinger's main loop
	r.thinger.Run(r.injector)
}
