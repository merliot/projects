//go:build !tinygo
// +build !tinygo

package nano_rp2040

import "github.com/merliot/merle"

func (n *nano_rp2040) ConnectAP(ssid, pass string) string {
	return ""
}

func (n *nano_rp2040) run(p *merle.Packet) {
	merle.RunForever(p)
}
