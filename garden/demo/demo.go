//go:build demo
// +build demo

package demo

import (
	"github.com/merliot/merle"
	"github.com/merliot/projects/garden"
)

type demo struct {
	Msg string
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
	}
}

func (d *demo) getState(p *merle.Packet) {
	d.Msg = merle.ReplyState
	p.Marshal(d).Reply()
}

func (d *demo) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:     merle.NoInit,
		merle.CmdRun:      merle.RunForever,
		merle.GetState:    d.getState,
	}
}

func (d *demo) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir:    "assets",
		HtmlTemplate: "templates/demo.html",
	}
}
