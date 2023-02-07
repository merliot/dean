package main

import (
	"embed"
	"fmt"
	"sync"
	"time"

	"github.com/merliot/dean"
)

var mu sync.Mutex

var dispatch = struct {
	Path string
}{}

var state = struct {
	Path string
	Foo  int
}{
	Path: "state",
}

var announce = struct {
	Path  string
	Model string
	Id    string
	Name  string
}{
	Path: "announce",
	Model: "foo",
}

type update struct {
	Path string
	Foo int
}

//go:embed index.html
var fs embed.FS

func handler(msg *dean.Msg) {
	fmt.Printf("%s\n", msg.String())

	mu.Lock()
	defer mu.Unlock()

	msg.Unmarshal(&dispatch)

	switch dispatch.Path {
	case "get/state":
		msg.Marshal(&state).Reply()
	case "update":
		var up update
		msg.Unmarshal(&up)
		state.Foo = up.Foo
		msg.Broadcast()
	}
}

func main () {

	var ann dean.Msg

	thing := dean.NewThing("user", "passwd", "THING", 10, handler, fs)

	thing.Addr = ":8080"
	go thing.ListenAndServe()

	thing.Dial("user", "passwd", "ws://localhost:8080/ws/", ann.Marshal(&announce))

	for {
		var msg dean.Msg
		var up = update{Path: "update", Foo: state.Foo+1,}
		thing.Inject(msg.Marshal(&up))
		time.Sleep(10 * time.Second)
	}
}
