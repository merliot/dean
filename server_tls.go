//go:build !tinygo

package dean

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func (s *Server) ServeTLS(host string) error {
	certsDir := "certs-" + s.thinger.Id()

	autocertManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
		Cache:      autocert.DirCache(certsDir),
	}

	s.TLSConfig = &tls.Config{
		GetCertificate: autocertManager.GetCertificate,
	}

	go http.ListenAndServe(":http", autocertManager.HTTPHandler(nil))

	s.Addr = ":https"
	return(s.ListenAndServeTLS("", ""))
}
