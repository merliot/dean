//go:build !tinygo

package dean

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"

	"golang.org/x/net/websocket"
)

// Server serves a Thing
type Server struct {
	thinger Thinger
	http.Server
	bus      *Bus
	injector *Injector
	subs     Subscribers

	makersMu   rwMutex
	makers     Makers
	modelsMu   rwMutex
	models     map[string]Thinger // model prototypes keyed by model
	thingsMu   rwMutex
	things     map[string]Thinger // keyed by id
	socketsMu  rwMutex
	sockets    map[Socketer]Thinger // keyed by socket
	handlersMu rwMutex
	handlers   map[string]http.HandlerFunc // keyed by path

	port   string
	user   string
	passwd string
}

// NewServer returns a server, serving the Thinger
func NewServer(thinger Thinger, user, passwd, port string) *Server {
	var s Server
	var id, _, _ = thinger.Identity()

	s.port = port
	s.user = user
	s.passwd = passwd

	fmt.Printf("    PORT:     %s\r\n", s.port)

	s.makers = Makers{}
	s.models = make(map[string]Thinger)
	s.things = make(map[string]Thinger)
	s.sockets = make(map[Socketer]Thinger)
	s.handlers = make(map[string]http.HandlerFunc)

	s.thinger = thinger
	s.thinger.SetOnline(true)

	s.subs = thinger.Subscribers()

	s.bus = NewBus("server bus", s.connect, s.disconnect)
	s.bus.Handle("", s.busHandle(thinger))
	s.injector = NewInjector("server injector", s.bus)

	mux := http.NewServeMux()
	s.Handler = mux

	mux.HandleFunc("/", s.root)

	if handler, ok := thinger.(http.Handler); ok {
		s.Handle("", handler)
	}
	s.HandleFunc("/server-state", s.serverState)
	s.HandleFunc("/ws/", s.serveWebSocket)
	s.HandleFunc("/ws/"+id+"/", s.serveWebSocket)

	return &s
}

func (s *Server) NewInjector(id string) *Injector {
	injector := NewInjector("thinger "+id+" injector", s.bus)
	injector.sock.SetTag(id)
	return injector
}

// RegisterModel registers a new Thing model
func (s *Server) RegisterModel(model string, maker ThingMaker) {
	s.makersMu.Lock()
	defer s.makersMu.Unlock()
	s.makers[model] = maker

	s.modelsMu.Lock()
	defer s.modelsMu.Unlock()
	s.models[model] = maker("proto", model, "proto")

	if handler, ok := s.models[model].(http.Handler); ok {
		s.Handle("/model/"+model+"/", http.StripPrefix("/model/"+model+"/", handler))
	}
}

// UnregisterModel unregisters the Thing model
func (s *Server) UnregisterModel(model string) {
	s.Unhandle("/model/" + model + "/")

	s.makersMu.Lock()
	defer s.makersMu.Unlock()
	delete(s.makers, model)

	s.modelsMu.Lock()
	defer s.modelsMu.Unlock()
	delete(s.models, model)
}

func (s *Server) Models() map[string]Thinger {
	return s.models
}

func (s *Server) dumpSockets(quote string) {
	fmt.Println("---- DUMP SOCKETS", quote)
	for k, v := range s.sockets {
		fmt.Printf("---- %s %s\n", k, v)
	}
}

func (s *Server) connect(socket Socketer) {
	fmt.Printf("\r\n*** CONNECT [%s]\r\n", socket)

	s.socketsMu.Lock()
	s.sockets[socket] = nil
	s.socketsMu.Unlock()
}

func (s *Server) handleAnnounce(pkt *Packet) {
	var ok bool
	var ann ThingMsgAnnounce
	pkt.Unmarshal(&ann)

	s.socketsMu.Lock()
	defer s.socketsMu.Unlock()

	socket := pkt.src
	if _, ok = s.sockets[socket]; !ok {
		fmt.Printf("Ignoring announcement: socket already disconnected: %s\r\n", socket)
		socket.Close()
		return
	}

	fmt.Printf("\r\n*** ANNOUNCE [%s] %s %s %s\r\n", socket, ann.Id, ann.Model, ann.Name)

	s.thingsMu.RLock()
	defer s.thingsMu.RUnlock()

	thinger, ok := s.things[ann.Id]
	if !ok {
		fmt.Printf("Ignoring annoucement: unknown thing Id %s\r\n", ann.Id)
		socket.Close()
		return
	}

	var id, model, name = thinger.Identity()

	if thinger.IsOnline() {
		fmt.Printf("Ignoring annoucement: thing already connected %s\r\n", id)
		socket.Close()
		return
	}

	if model != ann.Model {
		fmt.Printf("Ignoring annoucement: model doesn't match %s %s %s\r\n", id, model, ann.Model)
		socket.Close()
		return
	}

	if name != ann.Name {
		fmt.Printf("Ignoring annoucement: name doesn't match %s %s %s\r\n", id, name, ann.Name)
		socket.Close()
		return
	}

	thinger.SetOnline(true)
	socket.SetTag(id)

	s.sockets[socket] = thinger

	pkt.ClearPayload().SetPath("get/state").Reply()

	pkt.SetPath("connected").Marshal(&ThingMsgConnect{
		Id:    id,
		Model: model,
		Name:  name,
	})
	s.injector.Inject(pkt)

	// Notify other sockets with tag == id
	pkt.ClearPayload().SetPath("online")
	for sock := range s.sockets {
		if sock.Tag() == id && sock != socket {
			sock.Send(pkt)
		}
	}
}

func (s *Server) disconnect(socket Socketer) {
	fmt.Printf("\r\n*** DISCONNECT [%s]\r\n", socket)

	s.socketsMu.Lock()
	defer s.socketsMu.Unlock()

	thinger, ok := s.sockets[socket]
	if !ok {
		s.dumpSockets("deleting")
	}

	if thinger != nil {
		var pkt Packet
		var id, _, _ = thinger.Identity()

		thinger.SetOnline(false)

		pkt.SetPath("disconnected").Marshal(&ThingMsgDisconnect{Id: id})
		s.injector.Inject(&pkt)

		socket.SetTag("")

		// Notify other sockets with tag == id
		pkt.ClearPayload().SetPath("offline")
		for sock := range s.sockets {
			if sock.Tag() == id && sock != socket {
				sock.Send(&pkt)
			}
		}
	}

	delete(s.sockets, socket)
}

// CreateThing creates a new Thing based on model
func (s *Server) CreateThing(id, model, name string) (Thinger, error) {
	var pkt Packet

	if !ValidId(id) {
		return nil, fmt.Errorf("Invalid ID.  A valid ID is a non-empty string with only [a-z], [A-Z], [0-9], or underscore characters.")
	}
	if !ValidId(model) {
		return nil, fmt.Errorf("Invalid Model.  A valid Model is a non-empty string with only [a-z], [A-Z], [0-9], or underscore characters.")
	}
	if name == "" {
		return nil, fmt.Errorf("Invalid Name.  A valid Name is a non-empty string.")
	}

	s.thingsMu.Lock()
	defer s.thingsMu.Unlock()

	if s.things[id] != nil {
		return nil, fmt.Errorf("Thing ID '%s' already exists", id)
	}

	s.makersMu.RLock()
	defer s.makersMu.RUnlock()

	maker, ok := s.makers[model]
	if !ok {
		return nil, fmt.Errorf("Thing Model '%s' not registered", model)
	}

	thinger := maker(id, model, name)
	s.things[id] = thinger

	s.bus.Handle(id, s.busHandle(thinger))

	if handler, ok := thinger.(http.Handler); ok {
		s.Handle("/device/"+id+"/", http.StripPrefix("/device/"+id+"/", handler))
	}
	s.HandleFunc("/ws/"+id+"/", s.serveWebSocket)

	pkt.SetPath("created/thing").Marshal(&ThingMsgCreated{Id: id, Model: model, Name: name})
	s.injector.Inject(&pkt)

	return thinger, nil
}

// DeleteThing deletes a Thing given id
func (s *Server) DeleteThing(id string) error {
	var pkt Packet

	s.thingsMu.Lock()
	defer s.thingsMu.Unlock()

	if s.things[id] == nil {
		return fmt.Errorf("Thing ID '%s' not found", id)
	}

	s.Unhandle("/ws/" + id + "/")
	s.Unhandle("/device/" + id + "/")
	s.bus.Unhandle(id)

	delete(s.things, id)

	pkt.SetPath("deleted/thing").Marshal(&ThingMsgDeleted{Id: id})
	s.injector.Inject(&pkt)

	s.socketsMu.Lock()
	defer s.socketsMu.Unlock()

	for sock := range s.sockets {
		if sock.Tag() == id {
			sock.Close()
		}
	}

	return nil
}

// AdoptThing adds a thing to server
func (s *Server) AdoptThing(thinger Thinger) error {
	var pkt Packet
	var id, model, name = thinger.Identity()

	s.thingsMu.Lock()
	defer s.thingsMu.Unlock()

	if s.things[id] != nil {
		return fmt.Errorf("Thing ID '%s' already exists", id)
	}

	s.things[id] = thinger

	s.bus.Handle(id, s.busHandle(thinger))

	if handler, ok := thinger.(http.Handler); ok {
		s.Handle("/device/"+id+"/", http.StripPrefix("/device/"+id+"/", handler))
	}
	s.HandleFunc("/ws/"+id+"/", s.serveWebSocket)

	pkt.SetPath("adopted/thing").Marshal(&ThingMsgAdopted{Id: id, Model: model, Name: name})
	s.injector.Inject(&pkt)

	return nil
}

func (s *Server) busHandle(thinger Thinger) func(*Packet) {
	return func(pkt *Packet) {
		println(pkt.Path)

		switch pkt.Path {
		case "announce":
			go s.handleAnnounce(pkt)
			return
		case "get/state", "state":
			pkt.src.SetFlag(SocketFlagBcast)
		}

		if locker, ok := thinger.(Locker); ok {
			locker.Lock()
			defer locker.Unlock()
		}

		subs := thinger.Subscribers()
		if sub, ok := subs[pkt.Path]; ok {
			println("found sub for", pkt.Path)
			sub(pkt)
		}
	}
}

// MaxSocket sets the maximum number of sockets that can connect to the server
func (s *Server) MaxSockets(maxSockets int) {
	s.bus.MaxSockets(maxSockets)
}

// Dial connects server to other servers using a websocket.  url is
//
//	"ws://<server:port>/ws/&ping-period=<x>"   (HTTP)
//
// or
//
//	"wss://<server:port>/ws/&ping-period=<x>"  (HTTPS)
//
// ping-period (optional) in seconds to set ping-pong period on the websocket.
// Ping-pong is for detecting half-closed TCP sockets so both endpoints shut
// down the socket.

func (s *Server) Dial(url *url.URL, tries int) Socketer {
	ws := newWebSocket(url, "", s.bus)
	go ws.Dial(s.user, s.passwd, s.thinger.Announce(ws), tries)
	return ws
}

func (s *Server) Dials(urls string) {
	for _, u := range strings.Split(urls, ",") {
		if u == "" {
			continue
		}
		url, err := url.Parse(u)
		if err != nil {
			fmt.Printf("Error parsing URL: %s\r\n", err.Error())
			continue
		}
		switch url.Scheme {
		case "ws", "wss":
			s.Dial(url, -1)
		default:
			fmt.Println("Scheme must be ws or wss:", u)
		}
	}
}

func (s *Server) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	ws := newWebSocket(r.URL, r.RemoteAddr, s.bus)
	thingId := ws.getId()
	serverId, _, _ := s.thinger.Identity()
	if serverId != thingId {
		ws.SetTag(thingId)
	}
	serv := websocket.Server{Handler: websocket.Handler(ws.serve)}
	serv.ServeHTTP(w, r)
}

// Run the server
func (s *Server) Run() {

	// If we crash, put thinger in fail safe mode
	defer func() {
		if recover() != nil {
			s.thinger.FailSafe()
		}
	}()

	// Thinger is metal when run in server
	s.thinger.SetFlag(ThingFlagMetal)

	// Setup thinger
	s.thinger.Setup()

	// Start http server if valid listening port
	if s.port != "" {
		s.Addr = ":" + s.port
		go s.ListenAndServe()
	}

	// Run thinger's main loop
	s.thinger.Run(s.injector)
	id, _, name := s.thinger.Identity()
	fmt.Println("THINGER", name, id, "EXITED!")
}

func (s *Server) mux(w http.ResponseWriter, r *http.Request) {

	// Custom ServeMux.
	//
	// I tried using go1.22 http mux and it worked great,
	// but I couldn't figure out how to delete a route.  Need it for routes
	// like /ws/{id} or /device/{id} that come and go.

	_path := r.URL.Path

	// try exact match first
	if ok := s.runHandler(_path, w, r); ok {
		return
	}

	// try with removing last element from path
	dir, _ := path.Split(_path)
	if ok := s.runHandler(dir, w, r); ok {
		return
	}

	// redirect /device/{id}/* to child
	// e.g. /device/7667a30b-a3855397/js/temp.js
	parts := strings.Split(_path, "/")
	if len(parts) > 2 && parts[1] == "device" {
		id := parts[2]
		newpath := "/device/" + id + "/"
		if ok := s.runHandler(newpath, w, r); ok {
			return
		}
	}

	// redirect /model/{model}/* to model prototype
	// e.g. /model/temp/js/temp.js
	if len(parts) > 2 && parts[1] == "model" {
		model := parts[2]
		newpath := "/model/" + model + "/"
		if ok := s.runHandler(newpath, w, r); ok {
			return
		}
	}

	// everything else
	if ok := s.runHandler("", w, r); ok {
		return
	}

	// failed to find a match
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s not found", _path)
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {

	/*
		fmt.Printf("[%s] %s %s %s\n",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
		)
	*/

	// skip basic authentication if no user
	if s.user == "" {
		s.mux(w, r)
		return
	}

	ruser, rpasswd, ok := r.BasicAuth()

	if ok {
		userHash := sha256.Sum256([]byte(s.user))
		passHash := sha256.Sum256([]byte(s.passwd))
		ruserHash := sha256.Sum256([]byte(ruser))
		rpassHash := sha256.Sum256([]byte(rpasswd))

		// https://www.alexedwards.net/blog/basic-authentication-in-go
		userMatch := (subtle.ConstantTimeCompare(userHash[:], ruserHash[:]) == 1)
		passMatch := (subtle.ConstantTimeCompare(passHash[:], rpassHash[:]) == 1)

		if userMatch && passMatch {
			s.mux(w, r)
			return
		}
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (s *Server) runHandler(path string, w http.ResponseWriter, r *http.Request) bool {
	s.handlersMu.RLock()
	handler, ok := s.handlers[path]
	s.handlersMu.RUnlock()
	if ok {
		handler(w, r)
	}
	return ok
}

// HandleFunc registers an http handler func for path
func (s *Server) HandleFunc(path string, handler http.HandlerFunc) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.handlers[path] = handler
}

// HandleFunc registers an http handler for path
func (s *Server) Handle(path string, handler http.Handler) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.handlers[path] = handler.ServeHTTP
}

// Unhandle unregisters the http handler (or func) for path
func (s *Server) Unhandle(path string) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	delete(s.handlers, path)
}

var htmlBegin = `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="refresh" content="2">
  </head>
  <body>
    <pre>
      <code>`

var htmlEnd = `
      </code>
    </pre>
  </body>
</html>
`

func (s *Server) serverState(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintln(w, htmlBegin)

	fmt.Fprintln(w, "Thing: ", s.thinger)

	s.handlersMu.RLock()
	handlers := make([]string, 0, len(s.handlers))
	for key := range s.handlers {
		handlers = append(handlers, key)
	}
	s.handlersMu.RUnlock()
	sort.Strings(handlers)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Handlers")
	for _, handler := range handlers {
		fmt.Fprintf(w, "\t\"%s\"\n", handler)
	}

	s.socketsMu.Lock()
	sockets := make([]string, 0, len(s.sockets))
	for socket, thinger := range s.sockets {
		tag := socket.Tag()
		if tag == "" {
			tag = "{self}"
		}
		if thinger != nil {
			sockets = append(sockets, tag+", "+socket.String()+
				" "+thinger.String())
		} else {
			sockets = append(sockets, tag+", "+socket.String())
		}
	}
	s.socketsMu.Unlock()
	sort.Strings(sockets)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Sockets")
	for _, socket := range sockets {
		fmt.Fprintf(w, "\t%s\n", socket)
	}

	handlers = make([]string, 0, len(s.bus.handlers))
	for key := range s.bus.handlers {
		handlers = append(handlers, key)
	}
	sort.Strings(handlers)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Bus Handlers")
	for _, handler := range handlers {
		fmt.Fprintf(w, "\t\"%s\"\n", handler)
	}

	things := make([]string, 0, len(s.things))
	for _, thinger := range s.things {
		things = append(things, thinger.String())
	}
	sort.Strings(things)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Things")
	for _, thing := range things {
		fmt.Fprintf(w, "\t%s\n", thing)
	}

	fmt.Fprintln(w, htmlEnd)
}
