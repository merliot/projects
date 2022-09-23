//go:build wasm

package main

import (
	"github.com/merliot/merle"
	"github.com/merliot/projects/garden"
)

func main() {
	merle.NewWasm(garden.NewGarden()).Run()
}
