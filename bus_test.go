package dean

import (
	"testing"
	"time"
)

func TestNilConnect(t *testing.T) {
	bus := NewBus("test bus", nil, nil)
	sock := &socket{"test socket", "foo", 0, bus}
	bus.plugin(sock)
	bus.unplug(sock)
}

func TestConnect(t *testing.T) {
	i := 0
	connect := func(s Socketer) { i++ }
	disconnect := func(s Socketer) { i++ }
	bus := NewBus("test bus", connect, disconnect)
	sock := &socket{"test socket", "foo", 0, bus}
	bus.plugin(sock)
	i++
	bus.unplug(sock)
	if i != 3 {
		t.Error("Expected i == 3; i:", i)
	}
}

func TestNilHandler(t *testing.T) {
	defer func() { _ = recover() }()
	bus := NewBus("test bus", nil, nil)
	// should panic with nil handler
	bus.Handle("foo", nil)
	t.Errorf("did not panic")
}

func TestInvalidUnhandle(t *testing.T) {
	bus := NewBus("test bus", nil, nil)
	bus.Unhandle("foo")
}

func TestMultipleHandlers(t *testing.T) {
	foo := func(packet *Packet) { packet.src.SetTag("bar") }
	bar := func(packet *Packet) { packet.src.SetTag("baz") }
	baz := func(packet *Packet) { packet.src.SetTag("") }
	none := func(packet *Packet) { packet.src.SetTag("foo") }
	bus := NewBus("test bus", nil, nil)
	bus.Handle("foo", foo)
	bus.Handle("bar", bar)
	bus.Handle("baz", baz)
	bus.Handle("", none)
	sock := &socket{"test socket", "foo", 0, bus}
	packet := &Packet{bus, sock, nil}
	bus.receive(packet)
	bus.receive(packet)
	bus.receive(packet)
	if packet.src.Tag() != "" {
		t.Error("Expected \"\", got", packet.src.Tag())
	}
	bus.receive(packet)
	if packet.src.Tag() != "foo" {
		t.Error("Expected foo, got", packet.src.Tag())
	}
	bus.Unhandle("foo")
	bus.receive(packet)
	if packet.src.Tag() != "foo" {
		t.Error("Expected foo, got", packet.src.Tag())
	}
}

func TestMaxSocket(t *testing.T) {
	bus := NewBus("test bus", nil, nil)
	bus.MaxSockets(1)
	sock1 := &socket{"test socket 1", "foo", 0, bus}
	sock2 := &socket{"test socket 2", "foo", 0, bus}
	go func() { time.Sleep(time.Second); bus.unplug(sock1) }()
	bus.plugin(sock1)
	bus.plugin(sock2)
}

type testSocket struct {
	socket
	sent bool
}

func (s *testSocket) Send(packet *Packet) error {
	s.sent = true
	return nil
}

func TestBroadcast(t *testing.T) {
	bus := NewBus("test bus", nil, nil)
	sock1 := &testSocket{socket: socket{"test socket 1", "foo", SocketFlagBcast, bus}}
	sock2 := &testSocket{socket: socket{"test socket 2", "foo", SocketFlagBcast, bus}}
	sock3 := &testSocket{socket: socket{"test socket 3", "foo", 0, bus}}
	sock4 := &testSocket{socket: socket{"test socket 4", "foo", SocketFlagBcast, bus}}
	bus.plugin(sock1)
	bus.plugin(sock2)
	bus.plugin(sock3)
	bus.plugin(sock4)
	packet := &Packet{bus, sock1, nil}
	bus.broadcast(packet)
	if !(!sock1.sent && sock2.sent && !sock3.sent && sock4.sent) {
		t.Error("Broadcast failed")
	}
}
