// file: examples/garden/garden.go

package garden

import (
	"fmt"
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
	demoFlow    int
	// JSON exports
	Msg         string
	Demo        bool      `json:"-"`
	StartTime   string
	StartDays   [7]bool
	Gallons     float64
	GallonsGoal float64
	Running     bool
}

func NewGarden() *garden {
	return &garden{StartTime: "00:00", GallonsGoal: 1.0}
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
	g.Gallons = 0.0
	g.Running = false
	g.cmd = make(chan (int))

	if g.Demo {
		return
	}

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	g.relay = gpio.NewRelayDriver(adaptor, "31") // GPIO 6
	g.relay.Start()
	g.relay.Off()

	g.flowMeter = gpio.NewDirectPinDriver(adaptor, "7") // GPIO 4
	g.flowMeter.Start()
}

type msgUpdate struct {
	Msg   string
	Gallons     float64
	Running     bool
}

func (g *garden) update(p *merle.Packet) {
	var msg = msgUpdate{Msg: "Update"}
	g.Lock()
	if g.pulses >= g.pulsesGoal {
		g.Gallons = g.GallonsGoal
	} else {
		g.Gallons = float64(g.pulses) / pulsesPerGallon
	}
	msg.Gallons = g.Gallons
	msg.Running = g.Running
	g.Unlock()
	p.Marshal(&msg).Broadcast()
}

func (g *garden) pumpOn() {
	if !g.Demo {
		g.relay.On()
	}
}

func (g *garden) pumpOff() {
	if !g.Demo {
		g.relay.Off()
	}
}

func (g *garden) flow() (int, error) {
	if g.Demo {
		g.demoFlow++
		return g.demoFlow & 1, nil
	} else {
		return g.flowMeter.DigitalRead()
	}
}

func (g *garden) startWatering(p *merle.Packet) {
	prevVal, err := g.flow()
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
	g.pumpOn()

	sampler := time.NewTicker(5 * time.Millisecond)
	notify := time.NewTicker(time.Second)
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
			val, _ := g.flow()
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

	g.pumpOff()

	g.Lock()
	g.Running = false
	g.Unlock()

	g.update(p)
}

func (g *garden) run(p *merle.Packet) {
	waiting := false
	everySecond := time.NewTicker(time.Second)

	for {
		select {
		case _ = <-everySecond.C:
			now := time.Now()
			if g.StartDays[now.Weekday()] {
				hr, min, _ := now.Clock()
				hhmm := fmt.Sprintf("%d:%d", hr, min)
				if g.StartTime == hhmm {
					if !waiting {
						g.startWatering(p)
						// if we watered in less than a minute,
						// wait until next minute to check
						waiting = true
					}
				} else {
					waiting = false
				}
			}
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
