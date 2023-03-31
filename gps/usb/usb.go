package usb

import (
	"bufio"
	"log"
	"math"

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
		if rec.DataType() == nmea.TypeGLL {
			gll := rec.(nmea.GLL)
			if gll.Validity == "A" {
				u.Lat, u.Long = gll.Latitude, gll.Longitude
				dist := int(distance(u.Lat, u.Long, u.prevLat, u.prevLong) * 100.0) // cm
				println(dist)
				if dist > 20 {
					u.prevLat, u.prevLong = u.Lat, u.Long
					u.Path = "update"
					i.Inject(msg.Marshal(u))
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
  // must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}
