package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/merliot/merle"
	"tinygo.org/x/drivers/net/http"
	"tinygo.org/x/drivers/rtl8720dn"
)

var (
	ssid  string
	pass  string
	debug string
)

func init() {
	// 2 sec delay otherwise some printlns are missed at startup in serial output
	time.Sleep(2 * time.Second)

	adaptor := rtl8720dn.New(machine.UART3, machine.PB24,
		machine.PC24, machine.RTL8720D_CHIP_PU)
	adaptor.Debug(debug == "true")
	adaptor.Configure()

	http.UseDriver(adaptor)

	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	for err != nil {
		fmt.Println("Connect to AP failed: %w\r\n", err)
		time.Sleep(5 * time.Second)
	}

	ip, subnet, gateway, err := adaptor.GetIP()
	for err != nil {
		fmt.Println("Get IP failed: %w\r\n", err)
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("IP Address : %s\r\n", ip)
	fmt.Printf("Mask       : %s\r\n", subnet)
	fmt.Printf("Gateway    : %s\r\n", gateway)
}

type wio struct {
	Msg string
}

func NewWio() *wio {
	return new(wio)
}

func (w *wio) run(p *merle.Packet) {
	fmt.Printf("running...\r\n")
	select{}
}

func (w *wio) init(p *merle.Packet) {
	fmt.Printf("initing...\r\n")
}

func (w *wio) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit: w.init,
		merle.CmdRun:  w.run,
	}
}

func (w *wio) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
	}
}

func main() {
	thing := merle.NewThing(NewWio())
	thing.Cfg.PortPublic = 80
	thing.Run()
}
