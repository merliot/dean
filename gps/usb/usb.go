package usb

import (
	"bufio"
	"log"

	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/gps/nmea"
	"github.com/tarm/serial"
)

type Usb struct {
	*gps.Gps
	prevLat  float64
	prevLong float64
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS USB")
	return &Usb{
		Gps: gps.New(id, model, name).(*gps.Gps),
	}
}

func (u *Usb) Run(i *dean.Injector) {
	var msg dean.Msg

	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		println(scanner.Text())
		lat, long, err := nmea.ParseGLL(scanner.Text())
		if err != nil {
			println(err.Error())
			continue
		}
		u.Lat, u.Long = lat, long
		dist := int(gps.Distance(u.Lat, u.Long, u.prevLat, u.prevLong) * 100.0) // cm
		if dist < 20 {
			continue
		}
		u.prevLat, u.prevLong = u.Lat, u.Long
		u.Path = "update"
		i.Inject(msg.Marshal(u))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
