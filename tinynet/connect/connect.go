//go:build tinygo

package connect

import "github.com/merliot/dean/tinynet"

func init() {
	if ssid != "" {
		tinynet.NetConnect(ssid, pass)
	}
}
