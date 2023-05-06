//go:build !tinygo

package dean

import (
	"crypto/tls"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func (s *Server) serveTLS(host string) {
	certsDir := "certs-" + s.thinger.Id()

	autocertManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
		Cache:      autocert.DirCache(certsDir),
	}

	s.TLSConfig = &tls.Config{
		GetCertificate: autocertManager.GetCertificate,
	}

	log.Fatal(http.ListenAndServe(":http", autocertManager.HTTPHandler(nil)))
}

func (s *Server) ServeTLS(host string) error {
	go s.serveTLS(host)
	return s.ListenAndServeTLS("", "")
}
