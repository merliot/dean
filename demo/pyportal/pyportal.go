package pyportal

import (
	"embed"
	"net"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/dean/demo/base"
)

//go:embed css js index.html
var fs embed.FS

type Pyportal struct {
	*base.Base
	TempC int32
	Mac net.HardwareAddr
	Ip net.IP
}

func New(id, model, name string) dean.Thinger {
	println("NEW PYPORTAL")
	return &Pyportal{
		Base: base.New(id, model, name).(*base.Base),
	}
}

func (p *Pyportal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ServeFS(fs, w, r)
}
