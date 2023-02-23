package dean

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/net/websocket"
)

type Server struct {
	thinger Thinger
	http.Server
	*Bus
	*Injector
	handlers map[string]http.HandlerFunc
	clients  map[Socket]Thinger
	user     string
	passwd   string
}

func NewServer(thinger Thinger) *Server {
	var s Server

	s.handlers = make(map[string]http.HandlerFunc)
	s.clients = make(map[Socket]Thinger)

	s.thinger = thinger

	s.Bus = NewBus("server bus", s.connect, s.disconnect)
	s.Bus.Handle("", handle(thinger))
	s.Injector = NewInjector("server injector", s.Bus)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.root)
	s.Handler = mux

	s.Handle("/", thinger)
	s.HandleFunc("/state", s.state)
	s.HandleFunc("/ws/", s.Serve)
	s.HandleFunc("/ws/"+thinger.Id()+"/", s.Serve)

	return &s
}

func (s *Server) Register(socket Socket, client Thinger) {
	id := client.Id()
	s.clients[socket] = client
	socket.SetTag(id)
	s.Bus.Handle(id, handle(client))
	s.Handle("/"+id+"/", http.StripPrefix("/"+id+"/", client))
	s.HandleFunc("/ws/"+id+"/", s.Serve)
}

func (s *Server) connect(socket Socket) {
	println("*** CONNECT ", socket.Name(), socket)
	_, ok := s.clients[socket]
	if ok {
		println("ALREADY CONNECTED")
	}
	s.clients[socket] = nil
}

func (s *Server) disconnect(socket Socket) {
	if t := s.clients[socket]; t != nil {
		id := t.Id()
		s.Unhandle("/" + id + "/")
		s.Unhandle("/ws/" + id + "/")
		s.Bus.Unhandle(id)
		socket.SetTag("")
	}
	delete(s.clients, socket)
	println("*** DISCONNECT ", socket.Name())
}

func handle(thinger Thinger) func(*Msg) {
	return func(msg *Msg) {
		fmt.Printf("%s\n", msg.String())

		thinger.Lock()
		defer thinger.Unlock()

		var tmsg ThingMsg
		msg.Unmarshal(&tmsg)

		subs := thinger.Subscribers()
		if sub, ok := subs[tmsg.Path]; ok {
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

func (s *Server) Serve(w http.ResponseWriter, r *http.Request) {
	ws := NewWebSocket("websocket:"+r.Host, s.Bus)
	println("---- NewWebSocket --- ", ws)
	path := r.URL.Path
	println("path", path)
	parts := strings.Split(path, "/")
	if len(parts) == 2 {
		id := parts[1]
		println("tag", id)
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
	w.WriteHeader(http.StatusNotFound)
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

var html = `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="refresh" content="2">
  </head>
  <body>
    <pre>
      <code>
{{.}}
      </code>
    </pre>
  </body>
</html>
`

func (s *Server) state(w http.ResponseWriter, r *http.Request) {

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

	clients := make([]string, 0, len(s.clients))
	for client, thinger := range s.clients {
		if thinger != nil {
			clients = append(clients, client.Name() +
				" [Id: " + thinger.Id() +
				", Model: " + thinger.Model() +
				", Name: " + thinger.Name() + "]")
		} else {
			clients = append(clients, client.Name())
		}
	}
	sort.Strings(clients)
	f["Clients"] = clients

	var out bytes.Buffer
	b, _ := json.Marshal(f)
	json.Indent(&out, b, "", "\t")

	t, _ := template.New("foo").Parse(html)
	t.Execute(w, out.String())
}
