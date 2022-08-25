//go:build demo
// +build demo

package main

import (
	"flag"
	"log"

	"github.com/merliot/merle"
	"github.com/merliot/projects/garden"
	"github.com/merliot/projects/garden/demo"
)

func main() {
	garden := merle.NewThing(garden.NewGarden())

	garden.Cfg.Id = "garden"
	garden.Cfg.Model = "garden"
	garden.Cfg.Name = "garden01"
	garden.Cfg.PortPrivate = 7000
	garden.Cfg.MotherHost = "localhost"
	garden.Cfg.MotherUser = "merle"

	go garden.Run()

	demo := merle.NewThing(demo.NewDemo())

	demo.Cfg.Model = "garden_demo"
	demo.Cfg.Name = "garden_demo"
	demo.Cfg.PortPublic = 80
	demo.Cfg.PortPrivate = 6000

	flag.UintVar(&demo.Cfg.PortPublicTLS, "TLS", 0, "TLS port")
	flag.Parse()

	log.Fatalln(demo.Run())
}
