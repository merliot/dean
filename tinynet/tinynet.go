package tinynet

import "net"

func GetHardwareAddr() (net.HardwareAddr, error) {
	return netdev.GetHardwareAddr()
}

func GetIPAddr() (net.IP, error) {
	return netdev.GetIPAddr()
}
