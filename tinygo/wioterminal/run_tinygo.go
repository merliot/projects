//go:build tinygo
// +build tinygo

package wioterminal

import (
	"fmt"
	"machine"
	"time"

	"github.com/merliot/merle"
	"tinygo.org/x/drivers/net/http"
	"tinygo.org/x/drivers/rtl8720dn"
)

func init() {
	// 2 sec delay otherwise some printlns are missed at startup in serial output
	time.Sleep(2 * time.Second)
}

func (w *wio) Init(ssid, pass, debug string) (mac []byte, err error) {
	adaptor := rtl8720dn.New(machine.UART3, machine.PB24,
		machine.PC24, machine.RTL8720D_CHIP_PU)
	if debug == "true" {
		adaptor.Debug(true)
	}
	adaptor.Configure()

	http.UseDriver(adaptor)

	err = adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	if err != nil {
		fmt.Println("Connect to AP failed: %w\r\n", err)
		return []byte{}, err
	}

	ip, subnet, gateway, err := adaptor.GetIP()
	if err != nil {
		fmt.Println("Get IP failed: %w\r\n", err)
		return []byte{}, err
	}
	fmt.Printf("IP Address : %s\r\n", ip)
	fmt.Printf("Mask       : %s\r\n", subnet)
	fmt.Printf("Gateway    : %s\r\n", gateway)

	mac, err = adaptor.GetMac()
	if err != nil {
		fmt.Println("Get Mac failed: %w\r\n", err)
		return []byte{}, err
	}
	fmt.Printf("MAC Address : %s\r\n", mac)

	return mac, nil
}

func (w *wio) init(p *merle.Packet) {
	w.led = machine.LED
	w.led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	w.backlight = machine.LCD_BACKLIGHT
	w.backlight.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

func (w *wio) root(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "hello")
}

func (w *wio) run(p *merle.Packet) {
	http.HandleFunc("/", w.root)
	if err := http.ListenAndServe(":80", nil); err != nil {
		fmt.Printf("ListenAndServe error: %w", err)
	}
}
