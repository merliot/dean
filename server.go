package dean

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	thinger Thinger
	http.Server        `json:"-"`
	*Bus               `json:"-"`
	*Injector          `json:"-"`
	mux *http.ServeMux
	user string
	passwd string
}

func NewServer(thinger Thinger) *Server {
	var s Server
	s.thinger = thinger
	s.mux = http.NewServeMux()
	s.Handler = s.mux
	s.Bus = NewBus("server bus", thinger.Handler)
	s.Injector = NewInjector("server injector", s.Bus)
	return &s
}

func (s *Server) BasicAuth(user, passwd string) {
	s.user, s.passwd = user, passwd
}

func (s *Server) Dial(user, passwd, url string, announce *Msg) {
	ws := NewWebSocket("websocket:" + url, s.Bus)
	go ws.Dial(user, passwd, url, announce)
}

func (s *Server) Serve(w http.ResponseWriter, r *http.Request) {
	ws := NewWebSocket("websocket:" + r.Host, s.Bus)
	serv := websocket.Server{Handler: websocket.Handler(ws.serve)}
	serv.ServeHTTP(w, r)
}

func (s *Server) Run() {
	s.thinger.Run(s.Injector)
}

func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, s.basicAuth(handler))
}

func (s *Server) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {

		// skip basic authentication if no user
		if s.user == "" {
			next.ServeHTTP(writer, r)
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
				next.ServeHTTP(writer, r)
				return
			}
		}

		writer.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
	})
}
