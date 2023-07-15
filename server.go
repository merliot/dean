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
	bus        *Bus
	injector   *Injector
	subs       Subscribers
	handlersMu sync.RWMutex
	handlers   map[string]http.HandlerFunc
	socketsMu  sync.RWMutex
	sockets    map[Socketer]Thinger
	children   map[string]Thinger

	devicesMu  sync.RWMutex
	devices    map[string]Thinger
	makersMu   sync.RWMutex
	makers     Makers

	user       string
	passwd     string
}

func NewServer(thinger Thinger) *Server {
	var s Server

	s.handlers = make(map[string]http.HandlerFunc)
	s.sockets = make(map[Socketer]Thinger)
	s.children = make(map[string]Thinger)
	s.makers = dean.Makers{}

	s.thinger = thinger

	s.subs = thinger.Subscribers()

	s.bus = NewBus("server bus", s.connect, s.disconnect)
	s.bus.Handle("", s.busHandle(thinger))
	s.injector = NewInjector("server injector", s.bus)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.root)
	s.Handler = mux

	handler, ok := thinger.(http.Handler)
	if ok {
		s.Handle("", handler)
	}
	s.HandleFunc("/state", s.state)
	s.HandleFunc("/ws/", s.serveWebSocket)
	s.HandleFunc("/ws/"+thinger.Id()+"/", s.serveWebSocket)

	return &s
}

func (s *Server) RegisterModel(model string, maker dean.ThingMaker) {
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

func (s *Server) disconnect(socket Socketer) {

	if child := s.sockets[socket]; child != nil {
		id := child.Id()

		var msg Msg
		msg.Marshal(&ThingMsgDisconnect{"disconnected", id})
		s.injector.Inject(&msg)

		s.Unhandle("/ws/" + id + "/")
		s.bus.Unhandle(id)
		socket.SetTag("")

		//fmt.Printf("BEGIN closing other sockets\r\n")
		s.socketsMu.RLock()
		for sock := range s.sockets {
			if sock.Tag() == id && sock != socket {
				//fmt.Printf(">>>> closing %p\r\n", sock)
				sock.Close()
			}
		}
		s.socketsMu.RUnlock()
		//fmt.Printf("DONE closing other sockets\r\n")
	}

	s.socketsMu.Lock()
	delete(s.sockets, socket)
	s.socketsMu.Unlock()
	println("*** DISCONNECT", socket.Name())
}

func (s *Server) handleAnnounce(thinger Thinger, msg *Msg) {
	var ok bool
	var child Thinger
	var ann ThingMsgAnnounce
	msg.Unmarshal(&ann)

	id, model, name := ann.Id, ann.Model, ann.Name
	socket := msg.src

	s.socketsMu.RLock()
	for _, child := range s.sockets {
		if child != nil && child.Id() == id {
			println("CHILD ALREADY CONNECTED")
			socket.Close()
			s.socketsMu.RUnlock()
			return
		}
	}
	s.socketsMu.RUnlock()

	if child, ok = s.children[id]; !ok {
		maker, ok := thinger.(Maker)
		if !ok {
			// thinger is not a maker
			println("THINGER IS NOT A MAKER")
			return
		}
		child = maker.Make(id, model, name)
		if child == nil {
			// thinger couldn't make a child
			println("THINGER COULDN'T MAKE A THING")
			return
		}
		s.children[id] = child
		handler, ok := child.(http.Handler)
		if ok {
			s.Handle("/"+id+"/", http.StripPrefix("/"+id+"/", handler))
		}
	}

	socket.SetTag(id)

	s.socketsMu.Lock()
	s.sockets[socket] = child
	s.socketsMu.Unlock()
	//fmt.Printf(">>>> updated %p, %+v\r\n", socket, s.sockets)

	s.bus.Handle(id, s.busHandle(child))
	s.HandleFunc("/ws/"+id+"/", s.serveWebSocket)

	msg.Marshal(&ThingMsg{"get/state"}).Reply()
	msg.Marshal(&ThingMsgConnect{"connected", id, model, name})
	s.injector.Inject(msg)
}

func (s *Server) abandon(msg *Msg) {
	var orphan ThingMsgAbandon
	msg.Unmarshal(&orphan)

	s.Unhandle("/"+orphan.Id+"/")
	delete(s.children, orphan.Id)

	s.socketsMu.RLock()
	for socket, child := range s.sockets {
		if child != nil && child.Id() == orphan.Id {
			socket.Close()
			break
		}
	}
	s.socketsMu.RUnlock()
}

type ServerMsgModels struct {
	ThingMsg
	Models []string
}

func (s *Server) getModels(msg *Msg) {
	var models = ServerMsgModels{Path: "models"}
	s.makersMu.RLock()
	defer s.makersMu.RUnlock()
	for model in s.makers {
		models.Models = append(models.Models, model)
	}
	msg.Marshal(&models).Reply()
}

type ServerMsgCreate struct {
	ThingMsg
	Id string
	Model string
	Name string
	Err string
}

func (s *Server) _createDevice(id, model, name string)  error {
	if !ValidId(id) {
		return fmt.Errorf("Invalid ID.  A valid ID is a non-empty string with only [a-z], [A-Z], [0-9], or underscore characters.")
	}
	if !ValidId(model) {
		return fmt.Errorf("Invalid Model.  A valid Model is a non-empty string with only [a-z], [A-Z], [0-9], or underscore characters.")
	}
	if !ValidId(name) {
		return fmt.Errorf("Invalid Name.  A valid Name is a non-empty string with only [a-z], [A-Z], [0-9], or underscore characters.")
	}

	s.devicesMu.Lock()
	defer s.devicesMu.Unlock()

	if s.devices[id] != nil {
		return fmt.Errorf("Device ID '%s' already exists", id)
	}

	s.makersMu.RLock()
	defer s.makersMu.RUnlock()

	maker, ok := s.makers[model]
	if !ok {
		return fmt.Errorf("Device Model '%s' not registered", model)
	}

	s.devices[id] = maker(id, model, name)
	return nil
}

func (s *Server) createDevice(msg *Msg) {
	var create ServerMsgCreate
	msg.Unmarshal(&create)

	err := s._createDevice(create.Id, create.Model, create.Name)
	if err == nil {
		create.Path = "create/device/good"
		create.Err = ""
	} else {
		create.Path = "create/device/bad"
		create.Err = err.Error()
	}

	msg.Marshal(&create).Reply().Broadcast()
}


func (s *Server) busHandle(thinger Thinger) func(*Msg) {
	return func(msg *Msg) {
		fmt.Printf("Bus handle %s\r\n", msg.String())
		var rmsg ThingMsg

		msg.Unmarshal(&rmsg)

		switch rmsg.Path {
		case "get/models":
			s.getModels(msg)
			return
		case "create/device":
			s.createDevice(msg)
			return
		case "announce":
			go s.handleAnnounce(thinger, msg)
			return
		case "abandon":
			s.abandon(msg)
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
	var id string
	ws := newWebSocket("websocket:"+r.RemoteAddr, s.bus)
	id, ws.ping = ws.parsePath(r.URL.Path)
	if id != s.thinger.Id() {
		ws.SetTag(id)
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
	handler, ok := s.getHandler(_path)
	if ok {
		handler(w, r)
		return
	}

	// try with removing last element from path
	dir, _ := path.Split(_path)
	handler, ok = s.getHandler(dir)
	if ok {
		handler(w, r)
		return
	}

	// redirect /id/* to child
	parts := strings.Split(_path, "/")
	if len(parts) > 1 {
		id := parts[1]
		if _, ok := s.children[id]; ok {
			newpath := "/" + id + "/"
			handler, ok = s.getHandler(newpath)
			if ok {
				handler(w, r)
				return
			}
		}
	}

	// everything else
	handler, ok = s.getHandler("")
	if ok {
		handler(w, r)
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

func (s *Server) getHandler(path string) (http.HandlerFunc, bool) {
	s.handlersMu.RLock()
	defer s.handlersMu.RUnlock()
	handler, ok := s.handlers[path]
	return handler, ok
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

func (s *Server) state(w http.ResponseWriter, r *http.Request) {

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
	for socket, thing := range s.sockets {
		tag := socket.Tag()
		if tag == "" {
			tag = "{self}"
		}
		if thing != nil {
			sockets = append(sockets, tag+", "+socket.Name()+
				" "+thing.String())
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

	children := make([]string, 0, len(s.children))
	for _, child := range s.children {
		children = append(children, child.String())
	}
	sort.Strings(children)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Children")
	for _, child := range children {
		fmt.Fprintf(w, "\t%s\n", child)
	}

	fmt.Fprintln(w, htmlEnd)
}
