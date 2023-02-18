package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/hub"
)

var server *dean.Server
var clients = make(map[Socket]Thinger)

func announce(s dean.Socket, t Thinger) {
	p := path(t)
	clients[s] = t
	s.SetTag(t.ID())
	path := "/" + t.ID() + "/"
	server.Handle(path, http.StripPrefix(path, t.Serve))
	path := "/ws/" + t.ID() + "/"
	server.HandleFunc(path, server.Serve)
}

func connect(s dean.Socket) {
	clients[s] = nil
}

func disconnect(s dean.Socket) {
	if t := clients[s]; t != nil {
		path := "/" + t.ID() + "/"
		server.Unhandle(path)
		path := "/ws/" + t.ID() + "/"
		server.Unhandle(path)
		s.SetTag("")
	}
	delete(clients, s)
}

func main () {
	hub := hub.New("xxxxx", "hub", "hub1")

	hub.Register("gps", gps.New, announce)

	server = dean.NewServer(hub, connect, disconnect)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8081"

	server.HandleFunc("/", hub.Serve)
	server.HandleFunc("/ws/", server.Serve)

	server.ListenAndServe()
}
