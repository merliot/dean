package main

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/merliot/dean"
)

/*
var state struct {
	Foo int
}

func getState(m dean.Msg) {
	fmt.Printf("%v\n", m)
	m.Data = state
	m.Reply()
}
*/

//go:embed index.html
var fs embed.FS

func handler(msg *dean.Msg) {
	fmt.Printf("%v\n", msg)
	msg.Reply()
	//msg.Broadcast()
}

func main () {

	http.Handle("/", http.FileServer(http.FS(fs)))

	thing := dean.NewThing(10, handler)
	thing.Addr = ":8080"
	go thing.ListenAndServe()

	for {
		thing.Feed(dean.NewMsg([]byte("hello")))
		time.Sleep(time.Second)
	}
}
