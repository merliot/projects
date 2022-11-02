package nano_rp2040

import "github.com/merliot/merle"

//tinyjson:json
type nano_rp2040 struct {
	Msg     string
}

func New() *nano_rp2040 {
	return &nano_rp2040{Msg: merle.ReplyState}
}

func (n *nano_rp2040) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit: merle.NoInit,
		merle.CmdRun:  n.run,
	}
}

func (n *nano_rp2040) Assets() merle.ThingAssets {
	return merle.ThingAssets{}
}
