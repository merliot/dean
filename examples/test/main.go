package main

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/merliot/dean"
)

type dispatch struct {
	Path string
}

type state struct {
	dispatch
	Foo int
}

var s state

//go:embed index.html
var fs embed.FS

func handler(msg *dean.Msg) {
	fmt.Printf("%s\n", msg.String())

	var disp dispatch
	msg.Unmarshal(&disp)

	switch disp.Path {
	case "get/state":
		s.Path = "reply/state"
		msg.Marshal(&s)
		println(msg.String())
		msg.Reply()
	case "update":
		msg.Broadcast()
	}
}

type announce struct {
	dispatch
	Model string
}

func main () {

	var msg dean.Msg
	var ann = announce{
		Path: "announce",
		Model: "foo",
	}

	http.Handle("/", http.FileServer(http.FS(fs)))

	thing := dean.NewThing("THING", 10, handler)

	thing.Addr = ":8080"
	go thing.ListenAndServe()

	thing.Dial("ws://localhost:8080/ws/", msg.Marshal(&ann))

	for {
		s.Path = "update"
		msg.Marshal(&s)
		//thing.Broadcast(&msg)
		s.Foo++
		time.Sleep(time.Second)
	}
}
