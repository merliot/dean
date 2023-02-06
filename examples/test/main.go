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
	Path string
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
		msg.Marshal(&s).Reply()
	case "update":
		msg.Broadcast()
	}
}

type announce struct {
	Path string
	Model string
}

var ann = announce{
	Path: "announce",
	Model: "foo",
}

func main () {

	//var announce dean.Msg

	http.Handle("/", http.FileServer(http.FS(fs)))

	thing := dean.NewThing("THING", 10, handler)

	thing.Addr = ":8080"
	go thing.ListenAndServe()

	//thing.Dial("ws://localhost:8080/ws/", announce.Marshal(&ann))

	for {
		var update dean.Msg
		s.Path = "update"
		thing.Inject(update.Marshal(&s))
		s.Foo++
		time.Sleep(time.Second)
	}
}
