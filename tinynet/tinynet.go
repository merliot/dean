package tinynet

import "net"

func GetHardwareAddr() (net.HardwareAddr, error) {
	return netlink.GetHardwareAddr()
}

func GetIPAddr() (net.IP, error) {
	return netlink.GetIPAddr()
}
