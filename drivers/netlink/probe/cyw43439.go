//go:build pico

package probe

import (
	"github.com/soypat/cyw43439"
	//"github.com/merliot/dean/drivers/cyw43439"
	"github.com/merliot/dean/drivers/netdev"
	"github.com/merliot/dean/drivers/netlink"
	"github.com/merliot/dean/drivers/tcpip"
)

func Probe() (netlink.Netlinker, netdev.Netdever) {

	spi, cs, wlreg, irq := cyw43439.PicoWSpi(0)
	cyw43 := cyw43439.NewDevice(spi, cs, wlreg, irq, irq)

	stack := tcpip.NewStack(cyw43)
	netdev.UseNetdev(stack)

	return cyw43, stack
}
