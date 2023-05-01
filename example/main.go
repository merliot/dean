package main

import (
	"log"

	"github.com/merliot/dean"
)

type thing struct {
	dean.Thing
	dean.ThingMsg
}

func New(id, model, name string) dean.Thinger {
	return &thing{Thing: dean.NewThing(id, model, name)}
}

func main() {
	t := New("id", "model", "name")
	server := dean.NewServer(t)
	log.Fatal(server.ListenAndServe())
}
