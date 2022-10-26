// tinygo flash -target=wioterminal -serial usb cmd/wioterminal/main.go
//
// minicom -c on -D /dev/ttyACM0 -b 115200

package main

import (
	"log"
	"strings"

	"github.com/merliot/merle"
	"github.com/merliot/projects/tinygo/wioterminal"
)

var (
	ssid  string
	pass  string
	debug string
)

func main() {
	wio := wioterminal.New()
	mac, err := wio.Init(ssid, pass, debug)
	if err != nil {
		log.Fatalln(err)
	}

	thing := merle.NewThing(wio)

	thing.Cfg.Id = strings.Replace(string(mac), ":", "_", -1)
	thing.Cfg.MotherHost = "example.org"
	thing.Cfg.MotherUser = "foobar" // not used, but need not-empty string
	thing.Cfg.PortPrivate = 8080
	thing.Cfg.MotherPortPrivate = 8080

	log.Fatalln(thing.Run())
}
