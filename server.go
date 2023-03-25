package dean

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/merliot/dean/tinynet"
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
	thinger.SetReal()

	s.subs = thinger.Subscribers()

	s.Bus = NewBus("server bus", s.connect, s.disconnect)
	s.Bus.Handle("", s.busHandle(thinger))
	s.Injector = NewInjector("server injector", s.Bus)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.root)
	s.Handler = mux

	s.Handle("", thinger)
	s.HandleFunc("/state", s.state)
	s.HandleFunc("/ws/", s.serve)
	s.HandleFunc("/ws/"+thinger.Id()+"/", s.serve)

	return &s
}

func (s *Server) connect(socket Socket) {
	println("*** CONNECT ", socket.Name(), socket)
	_, ok := s.sockets[socket]
	if ok {
		panic("ALREADY CONNECTED")
	}
	s.sockets[socket] = nil
	fmt.Printf(">>>> added %p, %+v\n", socket, s.sockets)
}

func (s *Server) disconnect(socket Socket) {
	if thing := s.sockets[socket]; thing != nil {
		id := thing.Id()

		var msg Msg
		msg.Marshal(&ThingMsgDisconnect{"disconnected", id})
		s.Inject(&msg)

		s.Unhandle("/ws/" + id + "/")
		s.Bus.Unhandle(id)
		socket.SetTag("")

		fmt.Printf("BEGIN closing other sockets\n")
		for sock, _ := range s.sockets {
			if sock.Tag() == id && sock != socket {
				fmt.Printf(">>>> closing %p\n", sock)
				sock.Close()
			}
		}
		fmt.Printf("DONE closing other sockets\n")
	}
	fmt.Printf(">>>> before deleted %p, %+v\n", socket, s.sockets)
	delete(s.sockets, socket)
	fmt.Printf(">>>> after deleted %p, %+v\n", socket, s.sockets)
	println("*** DISCONNECT", socket.Name())
}

func (s *Server) handleAnnounce(thinger Thinger, msg *Msg) {
	var ok bool
	var thing Thinger
	var ann ThingMsgAnnounce
	msg.Unmarshal(&ann)

	id, model, name  := ann.Id, ann.Model, ann.Name

	if thing, ok = s.children[id]; !ok {
		maker, ok := thinger.(Maker)
		if !ok {
			// thinger is not a maker
			println("THINGER IS NOT A MAKER")
			return
		}
		thing = maker.Make(id, model, name)
		if thing == nil {
			// thinger couldn't make a thing
			println("THINGER COULDN'T MAKE A THING")
			return
		}
		s.children[id] = thing
		s.Handle("/"+id+"/", http.StripPrefix("/"+id+"/", thing))
	}

	socket := msg.src
	s.sockets[socket] = thing
	fmt.Printf(">>>> updated %p, %+v\n", socket, s.sockets)

	socket.SetTag(id)
	s.Bus.Handle(id, s.busHandle(thing))
	s.HandleFunc("/ws/"+id+"/", s.serve)

	msg.Marshal(&ThingMsg{"attached"}).Reply()
	msg.Marshal(&ThingMsgConnect{"connected", id, model, name})
	go s.Inject(msg)
}

func (s *Server) busHandle(thinger Thinger) func(*Msg) {
	return func(msg *Msg) {
		fmt.Printf("Bus handle %s\n", msg.String())

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

func (s *Server) Dial(user, passwd, url string, announce *Msg) {
	ws := NewWebSocket("websocket:"+url, s.Bus)
	go ws.Dial(user, passwd, url, announce)
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	ws := NewWebSocket("websocket:"+r.Host, s.Bus)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 4 {
		id := parts[2]
		if id != s.thinger.Id() {
			ws.SetTag(id)
		}
	}
	serv := websocket.Server{Handler: websocket.Handler(ws.serve)}
	serv.ServeHTTP(w, r)
}

func (s *Server) Run() {
	s.thinger.Run(s.Injector)
}

func (s *Server) mux(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	handler, ok := s.handlers[path]
	if ok {
		handler(w, r)
		return
	}
	handler, ok = s.handlers[""]
	if ok {
		handler(w, r)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s not found", path)
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
      <code>
`
var htmlEnd = `
      </code>
    </pre>
  </body>
</html>
`

func (s *Server) state(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintln(w, htmlBegin)

	fmt.Fprintln(w, "Id:   ", s.thinger.Id())
	fmt.Fprintln(w, "Model:", s.thinger.Model())
	fmt.Fprintln(w, "Name: ", s.thinger.Name())
	fmt.Fprintln(w, "User: ", s.user)

	// TODO finish this

	fmt.Fprintln(w, htmlEnd)

	/*
	f := map[string]any{
		"Id": s.thinger.Id(),
		"Model": s.thinger.Model(),
		"Name": s.thinger.Name(),
		"User": s.user,
		"Passwd": s.passwd,
	}

	handlers := make([]string, 0, len(s.handlers))
	for key := range s.handlers {
		handlers = append(handlers, key)
	}
	sort.Strings(handlers)
	f["Handlers"] = handlers

	sockets := make([]string, 0, len(s.sockets))
	for socket, thing := range s.sockets {
		if thing != nil {
			sockets = append(sockets, socket.Tag() +
				", " + socket.Name() +
				" [Id: " + thing.Id() +
				", Model: " + thing.Model() +
				", Name: " + thing.Name() + "]")
		} else {
			sockets = append(sockets, socket.Tag() +
				", " + socket.Name())
		}
	}
	sort.Strings(sockets)
	f["Sockets"] = sockets

	handlers = make([]string, 0, len(s.Bus.handlers))
	for key := range s.Bus.handlers {
		handlers = append(handlers, key)
	}
	sort.Strings(handlers)
	f["Bus Handlers"] = handlers

	children := make([]string, 0, len(s.children))
	for key := range s.children {
		children = append(children, key)
	}
	sort.Strings(children)
	f["Children"] = children

	var out bytes.Buffer
	b, _ := json.Marshal(f)
	json.Indent(&out, b, "", "\t")

	t, _ := template.New("foo").Parse(html)
	t.Execute(w, out.String())
	*/
}
