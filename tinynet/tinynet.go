package tinynet

import "net"

func GetHardwareAddr() (net.HardwareAddr, error) {
	return link.GetHardwareAddr()
}

func GetIPAddr() (net.IP, error) {
	return dev.GetIPAddr()
}
