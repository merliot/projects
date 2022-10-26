package wioterminal

import (
	"machine"

	"github.com/merliot/merle"
)

//tinyjson:json
type wio struct {
	Msg string
	led machine.Pin
	backlight machine.Pin
}

func New() *wio {
	return &wio{}
}

func (w *wio) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit: w.init,
		merle.CmdRun:  w.run,
	}
}

func (w *wio) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
	}
}
