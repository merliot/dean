//go:build challenger_rp2040

package probe

import (
	"machine"

	"github.com/merliot/dean/drivers/espat"
	"github.com/merliot/dean/drivers/netdev"
	"github.com/merliot/dean/drivers/netlink"
)

func Probe() (netlink.Netlinker, netdev.Netdever) {

	cfg := espat.Config{
		// UART
		Uart: machine.UART1,
		Tx:   machine.UART1_TX_PIN,
		Rx:   machine.UART1_RX_PIN,
	}

	esp := espat.NewDevice(&cfg)
	netdev.UseNetdev(esp)

	return esp, esp
}
