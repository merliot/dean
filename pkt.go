package dean

import (
	"encoding/json"
	"fmt"
)

// Packet is sent and received on a bus via a socket
type Packet struct {
	bus *Bus
	src Socketer
	message
}

func (p *Packet) String() string {
	return fmt.Sprintf("[%s] %s", p.src, p.message.String())
}

func (p *Packet) SetPath(path string) *Packet {
	p.Path = path
	return p
}

// Clear empties the packet's message payload
func (p *Packet) Clear() *Packet {
	p.Payload = []byte{}
	return p
}

// Reply sends the packet back to sender
func (p *Packet) Reply() *Packet {
	if p.src == nil {
		fmt.Printf("Can't reply to sender: source is nil\r\n")
		return p
	}
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
	fmt.Printf("Bcast %s\r\n", p)
	p.bus.broadcast(p)
	return p
}

// Unmarshal the packet message as JSON into v
func (p *Packet) Unmarshal(v any) *Packet {
	if err := json.Unmarshal(p.Payload, v); err != nil {
		fmt.Printf("JSON unmarshal error %s\r\n", err.Error())
	}
	return p
}

// Marshal the packet message as JSON from v
func (p *Packet) Marshal(v any) *Packet {
	var err error
	p.Payload, err = json.Marshal(v)
	if err != nil {
		fmt.Printf("JSON marshal error %s\r\n", err.Error())
	}
	return p
}
