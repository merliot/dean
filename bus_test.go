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
	foo := func(msg *Msg) { msg.src.SetTag("bar") }
	bar := func(msg *Msg) { msg.src.SetTag("baz") }
	baz := func(msg *Msg) { msg.src.SetTag("") }
	none := func(msg *Msg) { msg.src.SetTag("foo") }
	bus := NewBus("test bus", nil, nil)
	bus.Handle("foo", foo)
	bus.Handle("bar", bar)
	bus.Handle("baz", baz)
	bus.Handle("", none)
	sock := &socket{"test socket", "foo", 0, bus}
	msg := &Msg{bus, sock, nil}
	bus.receive(msg)
	bus.receive(msg)
	bus.receive(msg)
	if msg.src.Tag() != "" {
		t.Error("Expected \"\", got", msg.src.Tag())
	}
	bus.receive(msg)
	if msg.src.Tag() != "foo" {
		t.Error("Expected foo, got", msg.src.Tag())
	}
	bus.Unhandle("foo")
	bus.receive(msg)
	if msg.src.Tag() != "foo" {
		t.Error("Expected foo, got", msg.src.Tag())
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

func (s *testSocket) Send(msg *Msg) {
	s.sent = true
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
	msg := &Msg{bus, sock1, nil}
	bus.broadcast(msg)
	if !(!sock1.sent && sock2.sent && !sock3.sent && sock4.sent) {
		t.Error("Broadcast failed")
	}
}
