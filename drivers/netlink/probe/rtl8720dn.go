//go:build wioterminal

package probe

import (
	"machine"

	"github.com/merliot/dean/drivers/netdev"
	"github.com/merliot/dean/drivers/netlink"
	"github.com/merliot/dean/drivers/rtl8720dn"
)

func Probe() (netlink.Netlinker, netdev.Netdever) {

	cfg := rtl8720dn.Config{
		// Device
		En: machine.RTL8720D_CHIP_PU,
		// UART
		Uart:     machine.UART3,
		Tx:       machine.PB24,
		Rx:       machine.PC24,
		Baudrate: 614400,
	}

	rtl := rtl8720dn.New(&cfg)
	netdev.UseNetdev(rtl)

	return rtl, rtl
}
