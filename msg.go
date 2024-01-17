package dean

import (
	"encoding/json"
	"fmt"
)

// Msg is sent and received on a bus via a socket
type Msg struct {
	bus     *Bus
	src     Socketer
	payload []byte
}

// Bytes returns the msg payload
func (m *Msg) Bytes() []byte {
	return m.payload
}

func (m *Msg) String() string {
	return string(m.payload)
}

// Reply sends the msg back to sender.  The msg can be modified before calling
// Reply.
func (m *Msg) Reply() *Msg {
	if m.src == nil {
		fmt.Println("Can't reply to message: source is nil")
		return m
	}
	fmt.Println("Reply: src", m.src)
	m.src.Send(m)
	return m
}

// Broadcast the msg to all other matching-tagged sockets on the bus.  The
// source socket is excluded.
func (m *Msg) Broadcast() *Msg {
	if m.bus == nil {
		fmt.Println("Can't broadcast message: bus is nil")
		return m
	}
	fmt.Println("Broadcast: tag", m.src.Tag(), m)
	m.bus.broadcast(m)
	return m
}

// Unmarshal the msg payload as JSON into v
func (m *Msg) Unmarshal(v any) *Msg {
	err := json.Unmarshal(m.payload, v)
	if err != nil {
		fmt.Println("JSON unmarshal error", err.Error())
	}
	return m
}

// Marshal the msg payload as JSON from v
func (m *Msg) Marshal(v any) *Msg {
	var err error
	m.payload, err = json.Marshal(v)
	if err != nil {
		fmt.Println("JSON marshal error", err.Error())
	}
	return m
}
