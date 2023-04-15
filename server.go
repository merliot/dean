package dean

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"

	"golang.org/x/net/websocket"
)

type Server struct {
	thinger Thinger
	http.Server
	*Bus
	*Injector
	subs     Subscribers
	handlers map[string]http.HandlerFunc
	sockets  map[Socket]Thinger
	children map[string]Thinger
	user     string
	passwd   string
	msg      ThingMsg
}

func NewServer(thinger Thinger) *Server {
	var s Server

	s.handlers = make(map[string]http.HandlerFunc)
	s.sockets = make(map[Socket]Thinger)
	s.children = make(map[string]Thinger)

	s.thinger = thinger

	s.subs = thinger.Subscribers()

	s.Bus = NewBus("server bus", s.connect, s.disconnect)
	s.Bus.Handle("", s.busHandle(thinger))
	s.Injector = NewInjector("server injector", s.Bus)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.root)
	s.Handler = mux

	s.Handle("", thinger)
	s.HandleFunc("/state", s.state)
	s.HandleFunc("/ws/", s.serveWebSocket)
	s.HandleFunc("/ws/"+thinger.Id()+"/", s.serveWebSocket)

	return &s
}

func (s *Server) connect(socket Socket) {
	println("*** CONNECT ", socket.Name(), socket)
	_, ok := s.sockets[socket]
	if ok {
		panic("ALREADY CONNECTED")
	}
	s.sockets[socket] = nil
	fmt.Printf(">>>> added %p, %+v\r\n", socket, s.sockets)
}

func (s *Server) disconnect(socket Socket) {
	if child := s.sockets[socket]; child != nil {
		id := child.Id()

		var msg Msg
		msg.Marshal(&ThingMsgDisconnect{"disconnected", id})
		s.Inject(&msg)

		s.Unhandle("/ws/" + id + "/")
		s.Bus.Unhandle(id)
		socket.SetTag("")

		fmt.Printf("BEGIN closing other sockets\r\n")
		for sock := range s.sockets {
			if sock.Tag() == id && sock != socket {
				fmt.Printf(">>>> closing %p\r\n", sock)
				sock.Close()
			}
		}
		fmt.Printf("DONE closing other sockets\r\n")
	}
	fmt.Printf(">>>> before deleted %p, %+v\r\n", socket, s.sockets)
	delete(s.sockets, socket)
	fmt.Printf(">>>> after deleted %p, %+v\r\n", socket, s.sockets)
	println("*** DISCONNECT", socket.Name())
}

func (s *Server) handleAnnounce(thinger Thinger, msg *Msg) {
	var ok bool
	var child Thinger
	var ann ThingMsgAnnounce
	msg.Unmarshal(&ann)

	id, model, name := ann.Id, ann.Model, ann.Name

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
		s.Handle("/"+id+"/", http.StripPrefix("/"+id+"/", child))
	}

	socket := msg.src
	s.sockets[socket] = child
	fmt.Printf(">>>> updated %p, %+v\r\n", socket, s.sockets)

	socket.SetTag(id)
	s.Bus.Handle(id, s.busHandle(child))
	s.HandleFunc("/ws/"+id+"/", s.serveWebSocket)

	msg.Marshal(&ThingMsg{"attached"}).Reply()
	msg.Marshal(&ThingMsgConnect{"connected", id, model, name})
	go s.Inject(msg)
}

func (s *Server) busHandle(thinger Thinger) func(*Msg) {
	return func(msg *Msg) {
		fmt.Printf("Bus handle %s\r\n", msg.String())

		thinger.Lock()
		defer thinger.Unlock()

		msg.Unmarshal(&s.msg)

		switch s.msg.Path {
		case "announce":
			s.handleAnnounce(thinger, msg)
			return
		}

		subs := thinger.Subscribers()
		if sub, ok := subs[s.msg.Path]; ok {
			sub(msg)
		}
	}
}

func (s *Server) BasicAuth(user, passwd string) {
	s.user, s.passwd = user, passwd
}

func (s *Server) DialWebSocket(user, passwd, rawURL string, announce *Msg) {
	ws := NewWebSocket("websocket:"+rawURL, s.Bus)
	go ws.Dial(user, passwd, rawURL, announce)
}

const minPingMs = int(500) // 1/2 sec

func (s *Server) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	var id string
	ws := NewWebSocket("websocket:"+r.RemoteAddr, s.Bus)
	id, ws.ping = ws.parsePath(r.URL.Path)
	if id != s.thinger.Id() {
		ws.SetTag(id)
	}
	serv := websocket.Server{Handler: websocket.Handler(ws.serve)}
	serv.ServeHTTP(w, r)
}

func (s *Server) Run() {
	s.thinger.SetReal()
	s.thinger.Run(s.Injector)
}

func (s *Server) mux(w http.ResponseWriter, r *http.Request) {
	_path := r.URL.Path

	// try exact match first
	handler, ok := s.handlers[_path]
	if ok {
		handler(w, r)
		return
	}

	// Try with removing last element from path
	dir, _ := path.Split(_path)
	handler, ok = s.handlers[dir]
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
			handler, ok = s.handlers[newpath]
			if ok {
				handler(w, r)
				return
			}
		}
	}

	// everything else
	handler, ok = s.handlers[""]
	if ok {
		handler(w, r)
		return
	}
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

func (s *Server) HandleFunc(path string, handler http.HandlerFunc) {
	s.handlers[path] = handler
}

func (s *Server) Handle(path string, handler http.Handler) {
	s.handlers[path] = handler.ServeHTTP
}

func (s *Server) Unhandle(path string) {
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

	handlers := make([]string, 0, len(s.handlers))
	for key := range s.handlers {
		handlers = append(handlers, key)
	}
	sort.Strings(handlers)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Handlers")
	for _, handler := range handlers {
		fmt.Fprintf(w, "\t\"%s\"\n", handler)
	}

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
	sort.Strings(sockets)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Sockets")
	for _, socket := range sockets {
		fmt.Fprintf(w, "\t%s\n", socket)
	}

	handlers = make([]string, 0, len(s.Bus.handlers))
	for key := range s.Bus.handlers {
		handlers = append(handlers, key)
	}
	sort.Strings(handlers)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Bus Handlers")
	for _, handler := range handlers {
		fmt.Fprintf(w, "\t%s\n", handler)
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
