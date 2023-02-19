package main

import (
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/dean/gps"
	"github.com/merliot/dean/hub"
)

var server *dean.Server
var clients = make(map[dean.Socket]dean.Thinger)

func announce(s dean.Socket, t dean.Thinger) {
	id := t.Id()
	clients[s] = t
	s.SetTag(id)
	server.Bus.Handle(id, t.Subscribers)
	server.Handle("/"+id+"/", http.StripPrefix("/"+id+"/", t.Serve))
	server.HandleFunc("/ws/"+id+"/", server.Serve)
}

func connect(s dean.Socket) {
	clients[s] = nil
}

func disconnect(s dean.Socket) {
	if t := clients[s]; t != nil {
		id := t.Id()
		server.Unhandle("/" + id + "/")
		server.Unhandle("/ws/" + id + "/")
		server.Bus.Unhandle(id)
		s.SetTag("")
	}
	delete(clients, s)
}

func main() {
	h := hub.New("xxxxx", "hub", "hub1")

	h.Register("gps", hub.Factory{gps.New, announce})

	server = dean.NewServer(h, connect, disconnect)
	server.BasicAuth("user", "passwd")
	server.Addr = ":8081"

	server.HandleFunc("/", h.Serve)
	server.HandleFunc("/ws/", server.Serve)

	server.ListenAndServe()
}
