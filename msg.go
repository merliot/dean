package dean

import "encoding/json"

type Msg struct {
	bus     *Bus
	src     Socket
	payload []byte
}

func (m *Msg) Bytes() []byte {
	return m.payload
}

func (m *Msg) String() string {
	return string(m.payload)
}

func (m *Msg) Reply() {
	println("Reply: src", m.src.Name())
	m.src.Send(m)
}

func (m *Msg) Broadcast() {
	println("Broadcast: tag", m.src.Tag(), m.String())
	m.bus.broadcast(m)
}

func (m *Msg) Unmarshal(v any) *Msg {
	err := json.Unmarshal(m.payload, v)
	if err != nil {
		panic(err.Error())
	}
	return m
}

func (m *Msg) Marshal(v any) *Msg {
	var err error
	m.payload, err = json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	return m
}
