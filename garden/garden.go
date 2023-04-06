package garden

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/merliot/dean"
)

//go:embed css js index.html
var fs embed.FS

const (
	nZones int = 4
	// Sensor reads 450 pulses/Liter
	// 3.78541 Liters/Gallon
	pulsesPerGallon float64 = 450.0 * 3.78541
)

const (
	cmdStart int = iota
	cmdStop
)

type Zone struct {
	Name           string
	Index          int
	GallonsGoal    uint
	GallonsSoFar   uint
	TimeStarted    time.Time
	RunningTimeMax uint
	Running        bool
	cancel         chan bool
}

func (z *Zone) reset() {
	z.Running = false
	z.TimeStarted = time.Time{}
	z.GallonsSoFar = 0
}

func (z *Zone) stop() {
	select {
	case z.cancel <- true:
	default:
	}
}

type zoneMsg struct {
	Zone uint
}

type Garden struct {
	dean.Thing
	dean.ThingMsg
	SystemTime     time.Time
	StartTime      string
	StartDays      [7]bool
	Zones          []Zone
	timer          *time.Timer
	currGallons    uint
	*dean.Injector `json:"-"`
	PumpOn         func(*Zone) `json:"-"`
	PumpOff        func(*Zone) `json:"-"`
}

func New(id, model, name string) dean.Thinger {
	println("NEW GARDEN")
	var g Garden
	g.Thing = dean.NewThing(id, model, name)
	g.StartTime = "00:00"
	g.Zones = make([]Zone, nZones)
	for i := range g.Zones {
		g.Zones[i].Name = fmt.Sprintf("Zone %d", i+1)
		g.Zones[i].Index = i
		g.Zones[i].cancel = make(chan bool)
	}
	g.PumpOn  = func(z *Zone) { fmt.Println(z.Name, "pump ON") }
	g.PumpOff = func(z *Zone) { fmt.Println(z.Name, "pump OFF") }
	return &g
}

func (g *Garden) saveState(msg *dean.Msg) {
	msg.Unmarshal(g)
}

func (g *Garden) getState(msg *dean.Msg) {
	g.Path = "state"
	g.SystemTime = time.Now()
	msg.Marshal(g).Reply()
}

func (g *Garden) update(msg *dean.Msg) {
	msg.Unmarshal(g)
	dean.ThingStore(g)
	msg.Broadcast()
}

func (g *Garden) startTime(msg *dean.Msg) {
	g.update(msg)
	if g.IsReal() {
		g.schedule()
	}
}

func (g *Garden) getZone(msg *dean.Msg) uint {
	var z zoneMsg
	msg.Unmarshal(&z)
	fmt.Printf("%+v\n", z)
	return z.Zone
}

func (g *Garden) stopAllZones() {
	for zone := range g.Zones {
		g.Zones[zone].stop()
	}
}

func (g *Garden) startZone(msg *dean.Msg) {
	if g.IsReal() {
		g.stopAllZones()
		zone := g.getZone(msg)
		go g.runZone(&g.Zones[zone])
	}
	msg.Broadcast()
}

func (g *Garden) stopZone(msg *dean.Msg) {
	if g.IsReal() {
		zone := g.getZone(msg)
		g.Zones[zone].stop()
	}
	msg.Broadcast()
}

func (g *Garden) sendUpdate() {
	var msg dean.Msg
	g.Path = "update"
	g.Inject(msg.Marshal(g))
}

func (g *Garden) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     g.saveState,
		"get/state": g.getState,
		"attached":  g.getState,
		"update":    g.update,
		"starttime": g.startTime,
		"startzone": g.startZone,
		"stopzone":  g.stopZone,
	}
}

func (g *Garden) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.ServeFS(fs, w, r)
}

/*
func (g *Garden) init() {

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
*/

func (g *Garden) resetZones() {
	for i := range g.Zones {
		g.Zones[i].reset()
	}
	g.sendUpdate()
}

func (g *Garden) CurrentGallons() uint {
	g.currGallons++
	return g.currGallons
}

func (g *Garden) runZone(z *Zone) bool {

	var stopped = false

	if z.RunningTimeMax == 0 || z.GallonsGoal == 0 {
		return true
	}

	println("starting zone", z.Name)

	z.Running = true
	z.TimeStarted = time.Now()
	z.GallonsSoFar = 0
	g.sendUpdate()

	startGallons := g.CurrentGallons()

	g.PumpOn(z)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

run:
	for {
		select {
		case <-z.cancel:
			stopped = true
			break run
		case <-ticker.C:
			soFar := z.GallonsSoFar
			z.GallonsSoFar = g.CurrentGallons() - startGallons
			if z.GallonsSoFar != soFar {
				g.sendUpdate()
			}
			if z.GallonsSoFar >= z.GallonsGoal {
				break run
			}
			runningTime := time.Since(z.TimeStarted)
			if uint(runningTime.Minutes()) >= z.RunningTimeMax {
				break run
			}
		}
	}

	println("stop zone", z.Name)

	g.PumpOff(z)

	z.Running = false
	g.sendUpdate()

	return !stopped
}

func (g *Garden) runZones() {
	if g.StartDays[time.Now().Weekday()] {
		for i := range g.Zones {
			if !g.runZone(&g.Zones[i]) {
				break
			}
		}
	}
}

func (g *Garden) run() {
	println("Running")
	g.resetZones()
	g.runZones()
	g.schedule()
}

func (g *Garden) schedule() {
	hhmm, _ := time.ParseDuration(g.StartTime)
	now := time.Now()
	then := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	then = then.Add(hhmm)
	if now.After(then) {
		then = then.Add(24 * time.Hour) // add 24 hours to "then" if it's already passed today
	}
	wait := then.Sub(now)
	fmt.Printf("firing in %s\n", wait)
	if g.timer == nil {
		g.timer = time.AfterFunc(wait, g.run)
	} else {
		g.timer.Reset(wait)
	}
}

func (g *Garden) Run(i *dean.Injector) {

	g.Injector = i
	dean.ThingRestore(g)
	g.resetZones()
	g.schedule()

	select {}
}

/*
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
	// JSON exports
	Msg         string
	Now         time.Time
	StartTime   string
	StartDays   [7]bool
	GallonsGoal uint
}

func NewGarden() *garden {
	return &garden{
		loc:         time.UTC,
		GpioRelay:   31,      // GPIO 6
		GpioMeter:   7,       // GPIO 4
		StartTime:   "00:00",
		GallonsGoal: 1,
	}
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

type msgUpdate struct {
	Msg     string
	Gallons float64
	Running bool
}

func (g *garden) update(p *merle.Packet) {
	var msg = msgUpdate{Msg: "Update"}
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

type msgDay struct {
	Msg   string
	Day   uint
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
		"DateTime":       g.dateTime,
		"Update":         g.updateState,
		"Start":          g.start,
		"Stop":           g.stop,
		"Day":            g.day,
		"StartTime":      g.startTime,
		"GallonsGoal":    g.gallonsGoal,
	}
}
*/
