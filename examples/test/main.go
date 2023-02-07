package main

import (
	"embed"
	"fmt"
	"net/http"
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

	//var ann dean.Msg

	http.Handle("/", http.FileServer(http.FS(fs)))

	thing := dean.NewThing("THING", 10, handler)

	thing.Addr = ":8080"
	go thing.ListenAndServe()

	//thing.Dial("ws://localhost:8080/ws/", ann.Marshal(&announce))

	for {
		var msg dean.Msg
		var up = update{Path: "update", Foo: state.Foo+1,}
		thing.Inject(msg.Marshal(&up))
		time.Sleep(time.Second)
	}
}
