// file: examples/garden/garden.go

package garden

import (
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
	cmd         chan(int)
	pulses      int
	pulsesGoal  int
	Msg         string
	Gallons     float64
	GallonsGoal float64
}

func NewGarden() merle.Thinger {
	return &garden{GallonsGoal: 500.0}
}

func (g *garden) init(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	g.relay = gpio.NewRelayDriver(adaptor, "31") // GPIO 6
	g.relay.Start()
	g.relay.Off()

	g.flowMeter = gpio.NewDirectPinDriver(adaptor, "7") // GPIO 4
	g.flowMeter.Start()

	g.cmd = make(chan(int))
}

func (g *garden) update(p *merle.Packet) {
	g.Lock()
	g.Msg = merle.ReplyState
	g.Gallons = float64(g.pulses) / pulsesPerGallon
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
	g.Unlock()

	g.update(p)
	g.relay.On()

	sampler := time.NewTicker(5 * time.Millisecond)
	notify := time.NewTicker(5 * time.Second)
	defer sampler.Stop()
	defer notify.Stop()

	loop: for {
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

func (g *garden) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:    g.init,
		merle.CmdRun:     g.run,
		merle.GetState:   g.getState,
		merle.ReplyState: g.saveState,
		"Start":          g.start,
		"Stop":           g.stop,
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body style="background-color:lightblue">
		<button onclick="start()">Start</button>
		<button onclick="stop()">Stop</button>
		<div id="gallons">0</div>

		<script>
			var conn
			var online = false

			gallons = document.getElementById("gallons")

			function getState() {
				conn.send(JSON.stringify({Msg: "_GetState"}))
			}

			function getIdentity() {
				conn.send(JSON.stringify({Msg: "_GetIdentity"}))
			}

			function saveState(msg) {
				gallons.innerHTML = msg.Gallons
			}

			function showAll() {
			}

			function start() {
				conn.send(JSON.stringify({Msg: "Start"}))
			}

			function stop() {
				conn.send(JSON.stringify({Msg: "Stop"}))
			}

			function connect() {
				conn = new WebSocket("{{.WebSocket}}")

				conn.onopen = function(evt) {
					getIdentity()
				}

				conn.onclose = function(evt) {
					online = false
					showAll()
					setTimeout(connect, 1000)
				}

				conn.onerror = function(err) {
					conn.close()
				}

				conn.onmessage = function(evt) {
					msg = JSON.parse(evt.data)
					console.log('garden', msg)

					switch(msg.Msg) {
					case "_ReplyIdentity":
					case "_EventStatus":
						online = msg.Online
						getState()
					case "_ReplyState":
						saveState(msg)
						showAll()
						break
					}
				}
			}

			connect()
		</script>
	</body>
</html>`

func (g *garden) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		HtmlTemplateText: html,
	}
}
// file: examples/garden/garden.go
