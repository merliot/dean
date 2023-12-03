package dean

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
)

// Runner runs a Thing
type Runner struct {
	thinger  Thinger
	bus      *Bus
	injector *Injector

	user     string
	passwd   string
	dialURLs string
}

func NewRunner(thinger Thinger) *Runner {
	var r Runner

	r.user = os.Getenv("USER")
	r.passwd = os.Getenv("PASSWD")
	r.dialURLs = os.Getenv("DIAL_URLS")

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

// Dial connects runner to other servers
func (r *Runner) Dial() {
	for _, u := range strings.Split(r.dialURLs, ",") {
		purl, err := url.Parse(u)
		if err != nil {
			println("Error parsing URL:", err.Error())
			continue
		}
		switch purl.Scheme {
		case "ws", "wss":
			ws := newWebSocket(purl, r.bus)
			go ws.Dial(r.user, r.passwd, u, r.thinger.Announce())
		}
	}
}

// Run the main run loop
func (r *Runner) Run() {
	r.thinger.SetFlag(ThingFlagMetal)
	r.thinger.Run(r.injector)
}
