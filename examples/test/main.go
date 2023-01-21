package main

import (
	"github.com/merliot/dean"
)

func handler(m dean.Msg) {
}

func main () {
	d := dean.New()
	d.Handle("path/to/msg", handler)
	d.Dial("ws://localhost/ws")
	d.Serve(":80")
	//d.ServeTLS(":443")
}
