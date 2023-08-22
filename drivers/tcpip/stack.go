package tcpip

import (
	"github.com/merliot/dean/drivers/netlink"
)

type Stack struct {
	link netlink.Netlinker
}

func NewStack(link netlink.Netlinker) *Stack {
	s := Stack{link: link}
	s.link.RecvEthHandle(s.recvEth)
	return &s
}

func (s *Stack) recvEth(pkt []byte) error {
	println("recvEth", len(pkt))
	return nil
}
