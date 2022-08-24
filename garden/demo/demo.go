//go:build demo
// +build demo

package demo

import (
	"github.com/merliot/merle"
	"github.com/merliot/projects/garden"
)

type demo struct {
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

func (d *demo) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		"default": nil, // drop everything at the bridge level silently
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

func (d *demo) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:     merle.NoInit,
		merle.CmdRun:      merle.RunForever,
		merle.GetState:    d.getState,
		merle.EventStatus: d.update,
	}
}

func (d *demo) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir:    "assets",
		HtmlTemplate: "templates/demo.html",
	}
}
