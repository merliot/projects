// file: examples/garden/garden.go

package garden

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

// Sensor reads 450 pulses/Liter
// 3.78541 Liters/Gallon
var pulsesPerGallon float64 = 450.0 * 3.78541

const (
	cmdStart int = iota
	cmdStop
)

type garden struct {
	sync.Mutex
	relay       *gpio.RelayDriver
	flowMeter   *gpio.DirectPinDriver
	cmd         chan (int)
	pulses      int
	pulsesGoal  int
	Msg         string
	StartTime   string
	StartDays   [7]bool
	Gallons     float64
	GallonsGoal float64
	Running     bool
}

func NewGarden() merle.Thinger {
	return &garden{StartTime: "00:00", GallonsGoal: 500.0}
}

const store string = "store"

func (g *garden) store() {
	bytes, _ := json.Marshal(g)
	os.WriteFile(store, bytes, 0600)
}

func (g *garden) restore() {
	bytes, err := os.ReadFile(store)
	if err == nil {
		json.Unmarshal(bytes, g)
	}
}

func (g *garden) init(p *merle.Packet) {
	g.restore()

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	g.relay = gpio.NewRelayDriver(adaptor, "31") // GPIO 6
	g.relay.Start()
	g.relay.Off()

	g.flowMeter = gpio.NewDirectPinDriver(adaptor, "7") // GPIO 4
	g.flowMeter.Start()
	g.Gallons = 0.0

	g.cmd = make(chan (int))
}

func (g *garden) update(p *merle.Packet) {
	g.Lock()
	g.Msg = "Update"
	g.Gallons = float64(g.pulses) / pulsesPerGallon
	g.store()
	p.Marshal(g)
	g.Unlock()
	p.Broadcast()
}

func (g *garden) startWatering(p *merle.Packet) {
	prevVal, err := g.flowMeter.DigitalRead()
	if err != nil {
		println("ADC channel 0 read error:", err)
		return
	}

	g.Lock()
	g.pulses = 0
	g.pulsesGoal = int(g.GallonsGoal * pulsesPerGallon)
	g.Running = true
	g.Unlock()

	g.update(p)
	g.relay.On()

	sampler := time.NewTicker(5 * time.Millisecond)
	notify := time.NewTicker(5 * time.Second)
	defer sampler.Stop()
	defer notify.Stop()

loop:
	for {
		select {
		case cmd := <-g.cmd:
			switch cmd {
			case cmdStop:
				break loop
			}
		case _ = <-sampler.C:
			val, _ := g.flowMeter.DigitalRead()
			if val != prevVal {
				if val == 1 {
					g.Lock()
					g.pulses++
					g.Unlock()
				}
				prevVal = val
			}
		case _ = <-notify.C:
			g.update(p)
		}
		if g.pulses >= g.pulsesGoal {
			break loop
		}
	}

	g.relay.Off()

	g.Lock()
	g.Running = false
	g.Unlock()

	g.update(p)
}

func (g *garden) run(p *merle.Packet) {
	for {
		select {
		case cmd := <-g.cmd:
			switch cmd {
			case cmdStart:
				g.startWatering(p)
			}
		}
	}
}

func (g *garden) start(p *merle.Packet) {
	if p.IsThing() {
		g.cmd <- cmdStart
	} else {
		p.Broadcast()
	}
}

func (g *garden) stop(p *merle.Packet) {
	if p.IsThing() {
		g.cmd <- cmdStop
	} else {
		p.Broadcast()
	}
}

type msgDay struct {
	Msg   string
	Day   int
	State bool
}

func (g *garden) day(p *merle.Packet) {
	var msg msgDay
	p.Unmarshal(&msg)
	g.Lock()
	g.StartDays[msg.Day] = msg.State
	g.store()
	g.Unlock()
	p.Broadcast()
}

type msgStartTime struct {
	Msg   string
	Time  string
}

func (g *garden) startTime(p *merle.Packet) {
	var msg msgStartTime
	p.Unmarshal(&msg)
	g.Lock()
	g.StartTime = msg.Time
	g.store()
	g.Unlock()
	p.Broadcast()
}

func (g *garden) getState(p *merle.Packet) {
	g.Lock()
	g.Msg = merle.ReplyState
	p.Marshal(g)
	g.Unlock()
	p.Reply()
}

func (g *garden) saveState(p *merle.Packet) {
	g.Lock()
	p.Unmarshal(g)
	g.Unlock()
}

func (g *garden) updateState(p *merle.Packet) {
	g.saveState(p)
	p.Broadcast()
}

func (g *garden) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:    g.init,
		merle.CmdRun:     g.run,
		merle.GetState:   g.getState,
		merle.ReplyState: g.saveState,
		"Update":         g.updateState,
		"Start":          g.start,
		"Stop":           g.stop,
		"Day":            g.day,
		"StartTime":      g.startTime,
	}
}

func (g *garden) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir: "assets",
		HtmlTemplate: "templates/garden.html",
	}
}
