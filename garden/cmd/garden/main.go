package main

import (
	"flag"
	"log"

	"github.com/merliot/merle"
	"github.com/merliot/projects/garden"
)

func main() {
	thing := merle.NewThing(garden.NewGarden())

	thing.Cfg.Model = "garden"
	thing.Cfg.Name = "eden"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 6000

	flag.StringVar(&thing.Cfg.MotherHost, "rhost", "", "Remote host")
	flag.StringVar(&thing.Cfg.MotherUser, "ruser", "merle", "Remote user")
	flag.BoolVar(&thing.Cfg.IsPrime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")

	flag.Parse()

	log.Fatalln(thing.Run())
}
