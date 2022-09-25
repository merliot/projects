// file: projects/garden/garden.go

package garden

import (
	"fmt"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/merliot/merle"
	"github.com/merliot/projects/garden/msg"
	"github.com/merliot/projects/garden/state"
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
	dur         time.Duration
	loc         *time.Location
	Demo        bool      `json:"-"`
	GpioRelay   uint      `json:"-"`
	GpioMeter   uint      `json:"-"`
	state.State
}

func NewGarden() *garden {
	g := &garden{
		loc:         time.UTC,
		GpioRelay:   31,      // GPIO 6
		GpioMeter:   7,       // GPIO 4
	}
	g.StartTime = "00:00"
	g.GallonsGoal = 1
	return g
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

	relayPin := strconv.FormatUint(uint64(g.GpioRelay), 10)
	g.relay = gpio.NewRelayDriver(adaptor, relayPin)
	g.relay.Start()
	g.relay.Off()

	meterPin := strconv.FormatUint(uint64(g.GpioMeter), 10)
	g.flowMeter = gpio.NewDirectPinDriver(adaptor, meterPin)
	g.flowMeter.Start()
}

func (g *garden) update(p *merle.Packet) {
	var msg = msg.Update{Msg: "Update"}
	g.Lock()
	if g.pulses >= g.pulsesGoal {
		g.Gallons = float64(g.GallonsGoal)
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
	g.pulsesGoal = int(float64(g.GallonsGoal) * pulsesPerGallon)
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
	// Timer starts on 1 sec after next whole minute
	future := g.now().Truncate(time.Minute).
		Add(time.Minute).Add(time.Second)
	next := future.Sub(g.now())
	timer := time.NewTimer(next)

	for {
		select {
		case _ = <-timer.C:
			now := g.now()
			if g.StartDays[now.Weekday()] {
				hr, min, _ := now.Clock()
				hhmm := fmt.Sprintf("%02d:%02d", hr, min)
				if g.StartTime == hhmm {
					g.startWatering(p)
				}
			}
			// Timer starts on 1 sec after next whole minute
			future := g.now().Truncate(time.Minute).
				Add(time.Minute).Add(time.Second)
			next := future.Sub(g.now())
			timer = time.NewTimer(next)
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

func (g *garden) day(p *merle.Packet) {
	var msg msg.Day
	p.Unmarshal(&msg)
	g.Lock()
	g.StartDays[msg.Day] = msg.State
	g.store()
	g.Unlock()
	p.Broadcast()
}

func (g *garden) startTime(p *merle.Packet) {
	var msg msg.StartTime
	p.Unmarshal(&msg)
	g.Lock()
	g.StartTime = msg.Time
	g.store()
	g.Unlock()
	p.Broadcast()
}

type msgGallonsGoal struct {
	Msg         string
	GallonsGoal uint
}

func (g *garden) gallonsGoal(p *merle.Packet) {
	var msg msgGallonsGoal
	p.Unmarshal(&msg)
	g.Lock()
	g.GallonsGoal = msg.GallonsGoal
	g.store()
	g.Unlock()
	p.Broadcast()
}

func (g *garden) getState(p *merle.Packet) {
	g.Lock()
	g.Msg = merle.ReplyState
	g.Now = g.now()
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

type msgDateTime struct {
	Msg               string
	DateTime          time.Time
	ZoneOffsetMinutes int
}

func (g *garden) now() time.Time {
	// Represent time.Now() shifted by g.dur duration and in location g.loc
	return time.Now().Add(g.dur).In(g.loc)
}

func (g *garden) dateTime(p *merle.Packet) {
	// Only calculate time and time zone offsets from browser time and RPi
	// time once, on first browser contact.  So basically, system time will
	// be set to the time of the first browser contact.
	if g.loc == time.UTC {
		var msg msgDateTime
		p.Unmarshal(&msg)
		// g.dur is the time duration between RPi time and the time passed in
		// from the browser.  We'll add this duration time back to RPi time to
		// get browser time.
		g.dur = time.Now().Sub(msg.DateTime)
		// g.loc is the time.Location of the browser's time zone.
		g.loc = time.FixedZone("GardenTime", -msg.ZoneOffsetMinutes * 60)
	}

	p.Broadcast()
}

func (g *garden) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:    g.init,
		merle.CmdRun:     g.run,
		merle.GetState:   g.getState,
		merle.ReplyState: g.saveState,
		"DateTime":       g.dateTime,
		"Update":         g.updateState,
		"Start":          g.start,
		"Stop":           g.stop,
		"Day":            g.day,
		"StartTime":      g.startTime,
		"GallonsGoal":    g.gallonsGoal,
	}
}

func (g *garden) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir: "assets",
		HtmlTemplate: "templates/garden.html",
	}
}
