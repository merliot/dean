package main

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/merliot/dean"
)

var state struct {
	Foo int
}

func getState(m dean.Msg) {
	fmt.Printf("%v\n", m)
	m.Data = state
	m.Reply()
}

//go:embed index.html
var fs embed.FS

func main () {
	d := dean.New()
	d.Handle("get/state", getState)
	//d.Dial("ws://localhost:8080/ws")
	http.HandleFunc("/ws", d.Serve)
	http.Handle("/", http.FileServer(http.FS(fs)))
	http.ListenAndServe(":8080", nil)
}
