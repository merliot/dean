//go:build pyportal || nano_rp2040 || metro_m4_airlift || arduino_mkrwifi1010 || matrixportal_m4

package tinynet

import (
	"log"
	"time"

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var link netlink.Netlinker
var dev netdev.Netdever

func netConnect(ssid, pass string) error {

	link, dev = probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func init() {
	// wait a bit for serial
	time.Sleep(2 * time.Second)

	if err := netConnect(ssid, pass); err != nil {
		log.Fatal(err)
	}
}
