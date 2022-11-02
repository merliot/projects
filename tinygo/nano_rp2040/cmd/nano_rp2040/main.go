// file: examples/nano33/cmd/nano33/main.go

// tinygo flash -target=arduino-nano33 -serial usb -ldflags '-X "main.ssid=xxx" -X "main.pass=xxx"' cmd/nano33/main.go
//
// minicom -c on -D /dev/ttyACM0 -b 115200

package main

import (
	"log"
	"strings"

	"github.com/merliot/merle"
	"github.com/merliot/projects/tinygo/nano_rp2040"
)

var (
	ssid string
	pass string
)

func main() {
	nano := nano_rp2040.New()
	macAddr := nano.ConnectAP(ssid, pass)

	thing := merle.NewThing(nano)
	thing.Cfg.Id = strings.Replace(macAddr, ":", "_", -1)
	thing.Cfg.PortPublic = 80

	log.Fatalln(thing.Run())
}
