package dean

import (
	"encoding/json"
	"fmt"
)

// Msg is sent and received on a bus via a socket
type Msg struct {
	Tags    []string
	bus     *Bus
	src     Socketer
	payload []byte
}

func (m *Msg) tagInsert(tag string) {
	m.Tags = append([]string{tag}, m.Tags...)
}

func (m *Msg) tagStrip() string {
	if len(m.Tags) == 0 {
		return "" // No tags to remove
	}
	tag := m.Tags[0]
	m.Tags = m.Tags[1:]
	return tag
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
		fmt.Printf("Can't reply to message: source is nil\r\n")
		return m
	}
	fmt.Printf("Reply: src %s msg %s\r\n", m.src, m)
	m.src.Send(m)
	return m
}

// Broadcast the msg to all other matching-tagged sockets on the bus.  The
// source socket is excluded.
func (m *Msg) Broadcast() *Msg {
	if m.bus == nil {
		fmt.Printf("Can't broadcast message: bus is nil\r\n")
		return m
	}
	fmt.Printf("Broadcast: tag %s %s\r\n", m.src.Tag(), m)
	m.bus.broadcast(m)
	return m
}

// Unmarshal the msg payload as JSON into v
func (m *Msg) Unmarshal(v any) *Msg {
	err := json.Unmarshal(m.payload, v)
	if err != nil {
		fmt.Printf("JSON unmarshal error %s\r\n", err.Error())
	}
	return m
}

// Marshal the msg payload as JSON from v
func (m *Msg) Marshal(v any) *Msg {
	var err error
	m.payload, err = json.Marshal(v)
	if err != nil {
		fmt.Printf("JSON marshal error %s\r\n", err.Error())
	}
	return m
}
