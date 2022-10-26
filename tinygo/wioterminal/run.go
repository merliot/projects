//go:build !tinygo
// +build !tinygo

package wioterminal

import "github.com/merliot/merle"

func (w *wio) init(p *merle.Packet) {
}

func (w *wio) run(p *merle.Packet) {
	merle.RunForever(p)
}
