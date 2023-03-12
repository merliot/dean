//go:build demo
// +build demo

package demo

import (
	"github.com/merliot/merle"
	"github.com/merliot/projects/garden"
)

type demo struct {
	up      chan *merle.Packet
	Msg     string
	ChildId string
}

func NewDemo() merle.Thinger {
	return &demo{}
}

func (d *demo) BridgeThingers() merle.BridgeThingers {
	return merle.BridgeThingers{
		".*:garden:.*": func() merle.Thinger { return garden.NewGarden() },
	}
}

func (d *demo) upLevel(p *merle.Packet) {
	d.up <- p
}

func (d *demo) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		"Update":      d.upLevel,
		"Start":       d.upLevel,
		"Stop":        d.upLevel,
		"Day":         d.upLevel,
		"StartTime":   d.upLevel,
		"GallonsGoal": d.upLevel,
		"default":     nil, // drop everything else silently
	}
}

func (d *demo) getState(p *merle.Packet) {
	d.Msg = merle.ReplyState
	p.Marshal(d).Reply()
}

func (d *demo) update(p *merle.Packet) {
	var msg merle.MsgEventStatus
	p.Unmarshal(&msg)
	d.ChildId = ""
	if msg.Online {
		d.ChildId = msg.Id
	}
	p.Broadcast()
}

func (d *demo) init(p *merle.Packet) {
	d.up = make(chan *merle.Packet)
}

func (d *demo) run(p *merle.Packet) {
	for {
		select {
		case childP := <-d.up:
			childP.Copy(p)
			p.Broadcast()
		}
	}
}

func (d *demo) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:     d.init,
		merle.CmdRun:      d.run,
		merle.GetState:    d.getState,
		merle.EventStatus: d.update,
	}
}

func (d *demo) Assets() merle.ThingAssets {
	return merle.ThingAssets{
		AssetsDir:    "assets",
		HtmlTemplate: "templates/demo.html",
	}
}
