package dean

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Packet is sent and received on a bus via a socket
type Packet struct {
	bus     *Bus
	src     Socketer
	tags    string // tag chain
	message []byte // payload
}

// Bytes returns the packet message
func (p *Packet) Bytes() []byte {
	return p.message
}

func (p *Packet) String() string {
	return string(p.message)
}

// Tag pops the first tag off the tag chain and returns the tag
func (p *Packet) popTag() string {
	tags := strings.SplitN(p.tags, ".", 2)
	switch len(tags) {
	case 1:
		p.tags = ""
	case 2:
		p.tags = tags[1]
	}
	return tags[0]
}

// Reply sends the packet back to sender
func (p *Packet) Reply() *Packet {
	if p.src == nil {
		fmt.Printf("Can't reply to sender: source is nil\r\n")
		return p
	}
	fmt.Printf("Reply: src %s msg %s\r\n", p.src, p)
	p.src.Send(p)
	return p
}

// Broadcast the packet to all other matching-tagged sockets on the bus.  The
// source socket is excluded.
func (p *Packet) Broadcast() *Packet {
	if p.bus == nil {
		fmt.Printf("Can't broadcast packet: bus is nil\r\n")
		return p
	}
	fmt.Printf("Broadcast: tag %s %s\r\n", p.src.Tag(), p)
	p.bus.broadcast(p)
	return p
}

// Unmarshal the packet message as JSON into v
func (p *Packet) Unmarshal(v any) *Packet {
	err := json.Unmarshal(p.message, v)
	if err != nil {
		fmt.Printf("JSON unmarshal error %s\r\n", err.Error())
	}
	return p
}

// Marshal the packet message as JSON from v
func (p *Packet) Marshal(v any) *Packet {
	var err error
	p.message, err = json.Marshal(v)
	if err != nil {
		fmt.Printf("JSON marshal error %s\r\n", err.Error())
	}
	return p
}
