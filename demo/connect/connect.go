package connect

import (
	"embed"
	"net"
	"net/http"

	"github.com/merliot/dean"
)

//go:embed css js index.html
var fs embed.FS

type Connect struct {
	dean.Thing
	dean.ThingMsg
	CPUFreq float64
	Mac     string
	Ip      net.IP
	TempC   int32
	Lux     int32 // mlx (milliLux)
}

func New(id, model, name string) dean.Thinger {
	println("NEW CONNECT")
	return &Connect{
		Thing: dean.NewThing(id, model, name),
	}
}

func (c *Connect) saveState(msg *dean.Msg) {
	msg.Unmarshal(c)
}

func (c *Connect) getState(msg *dean.Msg) {
	c.Path = "state"
	msg.Marshal(c).Reply()
}

func (c *Connect) update(msg *dean.Msg) {
	msg.Unmarshal(c).Broadcast()
}

func (c *Connect) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     c.saveState,
		"get/state": c.getState,
		"attached":  c.getState,
		"update":    c.update,
	}
}

func (c *Connect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.ServeFS(fs, w, r)
}
