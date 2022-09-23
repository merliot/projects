//go:build wasm

package garden

import (
	"fmt"

	"github.com/merliot/merle"
)

func (g *garden) open(e *merle.Event) {
	fmt.Println("CmdOpen")
}

func (g *garden) WasmSubscribers() merle.WasmSubscribers {
	return merle.WasmSubscribers{
		merle.CmdOpen:    g.open,
	}
}
