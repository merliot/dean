package main

import (
	"github.com/merliot/dean"
)

func handler(m msg) {
}

func main () {
	d := dean.New()
	d.Handle("path/to/msg", handler)
	d.Dial("ws://localhost")
	d.Serve(":80")
	//d.ServeTLS(":443")
}
