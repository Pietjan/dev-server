package server

import (
	"fmt"
	"net/http"

	"github.com/pietjan/dev-server/pkg/server/router"
)

type Option = func(*http.Server)

func New(options ...func(*http.Server)) *http.Server {
	s := &http.Server{}

	for _, fn := range options {
		fn(s)
	}

	return s
}

func Router(options ...func(*http.ServeMux)) func(*http.Server) {
	return func(s *http.Server) {
		s.Handler = router.New(options...)
	}
}

func Port(port int) func(*http.Server) {
	return func(s *http.Server) {
		s.Addr = fmt.Sprintf("localhost:%d", port)
	}
}
