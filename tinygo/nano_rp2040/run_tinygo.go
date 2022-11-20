//go:build tinygo
// +build tinygo

package nano_rp2040

import (
	"fmt"
	"machine"
	"time"

	"github.com/merliot/merle"
	"tinygo.org/x/drivers/net/http"
	"tinygo.org/x/drivers/wifinina"
)

func init() {
	// 2 sec delay otherwise some printlns are missed at startup in serial output
	time.Sleep(2 * time.Second)
}

func (n *nano_rp2040) ConnectAP(ssid, pass string) string {
	// These are the default pins for the Arduino Nano33 IoT.
	spi := machine.NINA_SPI

	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       machine.NINA_SDO,
		SDI:       machine.NINA_SDI,
		SCK:       machine.NINA_SCK,
	})

	// This is the ESP chip that has the WIFININA firmware flashed on it
	adaptor := wifinina.New(spi,
		machine.NINA_CS,
		machine.NINA_ACK,
		machine.NINA_GPIO0,
		machine.NINA_RESETN)
	adaptor.Configure()

	http.UseDriver(adaptor)

	time.Sleep(2 * time.Second)
	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	for err != nil {
		fmt.Printf("Connect to AP failed: %s\r\n", err)
		time.Sleep(5 * time.Second)
	}

	ip, subnet, gateway, err := adaptor.GetIP()
	for err != nil {
		fmt.Println("Get IP failed: %s\r\n", err)
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("IP Address : %s\r\n", ip)
	fmt.Printf("Mask       : %s\r\n", subnet)
	fmt.Printf("Gateway    : %s\r\n", gateway)

	mac, err := adaptor.GetMACAddress()
	for err != nil {
		fmt.Println("Get MAC failed: %s\r\n", err)
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("MAC        : %s\r\n", mac.String())

	return mac.String()
}

func (n *nano_rp2040) run(p *merle.Packet) {
	fmt.Printf("running...\r\n")
	select{}
}
