//go:build pico

package probe

import (
	"github.com/soypat/cyw43439"
	//"tinygo.org/x/drivers/cyw43439"
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/tcpip"
)

func Probe() (netlink.Netlinker, netdev.Netdever) {

	spi, cs, wlreg, irq := cyw43439.PicoWSpi(0)
	cyw43 := cyw43439.NewDevice(spi, cs, wlreg, irq, irq)

	stack := tcpip.NewStack(cyw43)
	netdev.UseNetdev(stack)

	return cyw43, stack
}
