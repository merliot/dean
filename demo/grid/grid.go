package grid

import (
	"embed"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/dean/hub"
)

//go:embed css js index.html
var fs embed.FS

type Grid struct {
	*hub.Hub
}

func New(id, model, name string) dean.Thinger {
	println("NEW HUB GRID")
	return &Grid{
		Hub: hub.New(id, model, name).(*hub.Hub),
	}
}

func (g *Grid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.ServeFS(fs, w, r)
}
