//go:build tinygo

package main

import (
	"github.com/merliot/dean/tinynet"
)

var (
	ssid string
	pass string
)

func init() {
	tinynet.NetConnect(ssid, pass)
}

func main() {
}
