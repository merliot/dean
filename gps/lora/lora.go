package lora

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
)

type Lora struct {
	*gps.Gps
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS LORA")
	return &Lora{
		Gps: gps.New(id, model, name).(*gps.Gps),
	}
}
