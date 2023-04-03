package usb

import (
	"bufio"
	"log"

	"github.com/adrianmo/go-nmea"
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
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
		rec, err := nmea.Parse(scanner.Text())
		if err != nil {
			continue
		}
		if rec.DataType() != nmea.TypeGLL {
			continue
		}
		gll := rec.(nmea.GLL)
		if gll.Validity != "A" {
			continue
		}
		u.Lat, u.Long = gll.Latitude, gll.Longitude
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
