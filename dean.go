package dean

import (
)

type id uint64

type msg struct {
	src id
	path string
	data []byte
}

type handler func(m msg)

type dean struct {
	handlers map[string]handler
	bus chan(msg)
}

func New() {
	return &dean{
		handlers: make(map[string]handler)
		bus: make(chan(msg))
	}
}

func (d *dean) Dial(url string) error {
	return nil
}

func (d *dean) Handle(path string, h handler) {
	d.handlers[path] = h
}

func (d *dean) Serve(port string) error {
	return nil
}

func (d *dean) ServeTLS(port string) error {
	return nil
}
