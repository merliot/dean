//go:build !tinygo

package dean

import (
	"golang.org/x/crypto/acme/autocert"
)

func (s *Server) ServeTLS(host string) error {
	return s.Serve(autocert.NewListener(host))
}
