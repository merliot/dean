//go:build !tinygo

package dean

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"

	//sync "github.com/sasha-s/go-deadlock"
	"golang.org/x/net/websocket"
)

type Server struct {
	thinger Thinger
	http.Server
	bus      *Bus
	injector *Injector
	subs     Subscribers

	makersMu   sync.RWMutex
	makers     Makers
	thingsMu   sync.RWMutex
	things     map[string]Thinger // keyed by id
	socketsMu  sync.RWMutex
	sockets    map[Socketer]Thinger // keyed by socket
	handlersMu sync.RWMutex
	handlers   map[string]http.HandlerFunc // keyed by path

	user   string
	passwd string
}

func NewServer(thinger Thinger) *Server {
	var s Server
	var id, _, _ = thinger.Identity()

	s.makers = Makers{}
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
	mux.HandleFunc("/", s.root)
	s.Handler = mux

	if handler, ok := thinger.(http.Handler); ok {
		s.Handle("", handler)
	}
	s.HandleFunc("/server-state", s.serverState)
	s.HandleFunc("/ws/", s.serveWebSocket)
	s.HandleFunc("/ws/"+id+"/", s.serveWebSocket)

	return &s
}

func (s *Server) RegisterModel(model string, maker ThingMaker) {
	s.makersMu.Lock()
	defer s.makersMu.Unlock()
	s.makers[model] = maker
}

func (s *Server) UnregisterModel(model string) {
	s.makersMu.Lock()
	defer s.makersMu.Unlock()
	delete(s.makers, model)
}

func (s *Server) connect(socket Socketer) {
	println("*** CONNECT ", socket.Name(), socket)

	s.socketsMu.Lock()
	s.sockets[socket] = nil
	s.socketsMu.Unlock()
}

func (s *Server) handleAnnounce(msg *Msg) {
	var ok bool
	var ann ThingMsgAnnounce
	msg.Unmarshal(&ann)

	socket := msg.src

	println("*** ANNOUNCE ", socket.Name(), ann.Id, ann.Model, ann.Name)

	s.thingsMu.RLock()
	defer s.thingsMu.RUnlock()

	thinger, ok := s.things[ann.Id]
	if !ok {
		fmt.Println("Ignoring annoucement: unknown thing Id", ann.Id)
		socket.Close()
		return
	}

	var id, model, name = thinger.Identity()

	if model != ann.Model {
		fmt.Println("Ignoring annoucement: model doesn't match", id, model, ann.Model)
		socket.Close()
		return
	}

	if name != ann.Name {
		fmt.Println("Ignoring annoucement: name doesn't match", id, name, ann.Name)
		socket.Close()
		return
	}

	if thinger.IsOnline() {
		fmt.Println("Ignoring annoucement: thing already connected", id)
		socket.Close()
		return
	}

	thinger.SetOnline(true)
	socket.SetTag(id)

	s.socketsMu.Lock()
	defer s.socketsMu.Unlock()

	s.sockets[socket] = thinger

	msg.Marshal(&ThingMsg{"get/state"}).Reply()

	msg.Marshal(&ThingMsgConnect{"connected", id, model, name})
	s.injector.Inject(msg)

	// Notify other sockets with tag == id
	msg.Marshal(&ThingMsg{"online"})
	for sock := range s.sockets {
		if sock.Tag() == id && sock != socket {
			sock.Send(msg)
		}
	}
}

func (s *Server) disconnect(socket Socketer) {
	println("*** DISCONNECT", socket.Name())

	s.socketsMu.Lock()
	defer s.socketsMu.Unlock()

	thinger := s.sockets[socket]

	if thinger != nil {
		var msg Msg
		var id, _, _ = thinger.Identity()

		thinger.SetOnline(false)

		msg.Marshal(&ThingMsgDisconnect{"disconnected", id})
		s.injector.Inject(&msg)

		socket.SetTag("")

		// Notify other sockets with tag == id
		msg.Marshal(&ThingMsg{"offline"})
		for sock := range s.sockets {
			if sock.Tag() == id && sock != socket {
				sock.Send(&msg)
			}
		}
	}

	delete(s.sockets, socket)
}

func (s *Server) GetModels() []string {
	var models []string
	s.makersMu.RLock()
	defer s.makersMu.RUnlock()
	for model := range s.makers {
		models = append(models, model)
	}
	return models
}

func (s *Server) CreateThing(id, model, name string) (Thinger, error) {
	var msg Msg

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
		s.Handle("/"+id+"/", http.StripPrefix("/"+id+"/", handler))
	}
	s.HandleFunc("/ws/"+id+"/", s.serveWebSocket)

	msg.Marshal(&ThingMsgCreated{Path: "created/thing", Id: id, Model: model, Name: name})
	s.injector.Inject(&msg)

	return thinger, nil
}

func (s *Server) DeleteThing(id string) error {
	var msg Msg

	s.thingsMu.Lock()
	defer s.thingsMu.Unlock()

	if s.things[id] == nil {
		return fmt.Errorf("Thing ID '%s' not found", id)
	}

	s.Unhandle("/ws/" + id + "/")
	s.Unhandle("/" + id + "/")
	s.bus.Unhandle(id)

	delete(s.things, id)

	msg.Marshal(&ThingMsgDeleted{Path: "deleted/thing", Id: id})
	s.injector.Inject(&msg)

	s.socketsMu.Lock()
	defer s.socketsMu.Unlock()

	for sock := range s.sockets {
		if sock.Tag() == id {
			sock.Close()
		}
	}

	return nil
}

func (s *Server) busHandle(thinger Thinger) func(*Msg) {
	return func(msg *Msg) {
		fmt.Printf("Bus handle %s\r\n", msg.String())
		var rmsg ThingMsg

		msg.Unmarshal(&rmsg)

		switch rmsg.Path {
		case "announce":
			go s.handleAnnounce(msg)
			return
		case "get/state", "state":
			msg.src.SetFlag(SocketFlagBcast)
		}

		if locker, ok := thinger.(sync.Locker); ok {
			locker.Lock()
			defer locker.Unlock()
		}

		subs := thinger.Subscribers()
		if sub, ok := subs[rmsg.Path]; ok {
			sub(msg)
		}
	}
}

func (s *Server) MaxSockets(maxSockets int) {
	s.bus.MaxSockets(maxSockets)
}

func (s *Server) BasicAuth(user, passwd string) {
	s.user, s.passwd = user, passwd
}

func (s *Server) DialWebSocket(user, passwd, rawURL string, announce *Msg) {
	ws := newWebSocket("websocket:"+rawURL, s.bus)
	go ws.Dial(user, passwd, rawURL, announce)
}

func (s *Server) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	var thingId string
	var serverId, _, _ = s.thinger.Identity()
	ws := newWebSocket("websocket:"+r.RemoteAddr, s.bus)
	thingId, ws.ping = ws.parsePath(r.URL.Path)
	if serverId != thingId {
		ws.SetTag(thingId)
	}
	serv := websocket.Server{Handler: websocket.Handler(ws.serve)}
	serv.ServeHTTP(w, r)
}

func (s *Server) Run() {
	s.thinger.SetFlag(ThingFlagMetal)
	s.thinger.Run(s.injector)
}

func (s *Server) mux(w http.ResponseWriter, r *http.Request) {
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

	// redirect /id/* to child
	parts := strings.Split(_path, "/")
	if len(parts) > 1 {
		id := parts[1]
		newpath := "/" + id + "/"
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

func (s *Server) HandleFunc(path string, handler http.HandlerFunc) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.handlers[path] = handler
}

func (s *Server) Handle(path string, handler http.Handler) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.handlers[path] = handler.ServeHTTP
}

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

	fmt.Fprintln(w, "Thing: ", s.thinger.String())

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
			sockets = append(sockets, tag+", "+socket.Name()+
				" "+thinger.String())
		} else {
			sockets = append(sockets, tag+", "+socket.Name())
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
