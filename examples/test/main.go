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

	bus := dean.NewBus(10, handler)

	client := dean.NewWebSocket(bus)
	go client.Dial("ws://localhost:8080/ws")

	server := dean.NewWebSocketServer(bus)
	http.HandleFunc("/ws", server.Serve)
	http.Handle("/", http.FileServer(http.FS(fs)))
	go http.ListenAndServe(":8080", nil)

	host := dean.NewInjector(bus)

	for {
		host.Inject(dean.NewMsg([]byte("hello")))
		time.Sleep(time.Second)
	}
}
