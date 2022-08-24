//go:build demo
// +build demo

package main

import (
	"flag"
	"log"

	"github.com/merliot/merle"
	"github.com/merliot/projects/garden/demo"
)

func main() {
	thing := merle.NewThing(demo.NewDemo())

	thing.Cfg.Model = "garden_demo"
	thing.Cfg.Name = "garden_demo"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 6000

	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")

	flag.Parse()

	log.Fatalln(thing.Run())
}
