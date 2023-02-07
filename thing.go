package dean

import (
	"embed"
	"net/http"

	"golang.org/x/net/websocket"
)

type Thing struct {
	http.Server
	name     string
	bus      *Bus
	injector *injector
}

func NewThing(name string, maxSockets int, handler func(*Msg)) *Thing {
	bus := NewBus("thing " + name, maxSockets, handler)
	t := &Thing{
		name:     name,
		bus:      bus,
		injector: NewInjector("injector", bus),
	}
	http.HandleFunc("/ws/", t.basicAuth(t.serve))
	return t
}

func (t *Thing) Dial(url string, announce *Msg) {
	s := NewWebSocket("websocket:" + url, t.bus)
	go s.Dial(url, announce)
}

func (t *Thing) basicAuth(user, passwd string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {

		// skip basic authentication if no user
		if user == "" {
			next.ServeHTTP(writer, r)
			return
		}

		ruser, rpasswd, ok := r.BasicAuth()

		if ok {
			userHash := sha256.Sum256([]byte(user))
			passHash := sha256.Sum256([]byte(passwd))
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

func (t *Thing) serve(w http.ResponseWriter, r *http.Request) {
	ws := NewWebSocket("websocket:" + r.Host, t.bus)
	s := websocket.Server{Handler: websocket.Handler(ws.serve)}
	s.ServeHTTP(w, r)
}

func (t *Thing) Inject(msg *Msg) {
	t.injector.Inject(msg)
}

/*
type Thinger interface {
	Handler(*Msg)
	FS() embed.FS
}
*/
