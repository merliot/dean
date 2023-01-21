package main

import (
	"fmt"
	"net/http"

	"github.com/merliot/dean"
)

func handler(m dean.Msg) {
	fmt.Printf("%v\n", m)
	m.Data = []byte("bar")
	m.Reply()
}

func main () {
	d := dean.New()
	d.Handle("path/to/msg", handler)
	d.Dial("ws://localhost:8080/ws")
	http.HandleFunc("/ws", d.Serve)
	http.ListenAndServe(":8080", nil)
}
