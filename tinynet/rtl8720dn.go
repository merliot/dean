//go:build wioterminal

package tinynet

import (
	"log"
	"machine"
	"time"

	"tinygo.org/x/drivers/rtl8720dn"
)

var cfg = rtl8720dn.Config{
	// WiFi AP credentials
	Ssid:       ssid,
	Passphrase: pass,
	// Device
	En: machine.RTL8720D_CHIP_PU,
	// UART
	Uart:     machine.UART3,
	Tx:       machine.PB24,
	Rx:       machine.PC24,
	Baudrate: 614400,
	// Watchdog (set to 0 to disable)
	WatchdogTimeo: time.Duration(20 * time.Second),
}

var netdev = rtl8720dn.New(&cfg)

func init() {
	// wait a bit for serial
	time.Sleep(2 * time.Second)

	if err := netdev.NetConnect(); err != nil {
		log.Fatal(err)
	}
}
