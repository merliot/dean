// This example is an HTTP server serving up a static file system
//
// Note: It may be necessary to increase the stack size when using "net/http".
// Use the -stack-size=4KB command line option.

//go:build pyportal || nano_rp2040 || metro_m4_airlift || arduino_mkrwifi1010 || matrixportal_m4 || wioterminal

package main

import (
	"embed"
	"log"
	"net/http"
	"time"

	"github.com/merliot/dean/drivers/netlink"
	"github.com/merliot/dean/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	port string = ":80"
)

//go:embed index.html main.go images
var fs embed.FS

func main() {
	// wait a bit for console
	time.Sleep(2 * time.Second)

	link, _ := probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		log.Fatal(err)
	}

	hfs := http.FileServer(http.FS(fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hfs.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
