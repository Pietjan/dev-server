package router

import (
	"fmt"
	"net/http"

	"github.com/pietjan/dev-server/pkg/server/proxy"
	"github.com/pietjan/dev-server/pkg/server/ws"
	"github.com/pietjan/events"
)

type Option = func(*http.ServeMux)

func New(options ...func(*http.ServeMux)) http.Handler {
	router := http.NewServeMux()

	for _, fn := range options {
		fn(router)
	}

	return router
}

func ProxyTarget(port int) func(*http.ServeMux) {
	return func(router *http.ServeMux) {
		router.Handle("/", proxy.New(
			proxy.Target(fmt.Sprintf("http://localhost:%d", port)),
		))
	}
}

func WS(processor *events.Processor) func(*http.ServeMux) {
	return func(router *http.ServeMux) {
		router.HandleFunc("GET /__dev-server/ws", ws.Handler(processor))
	}
}

func LiveReloadWS(script string) func(*http.ServeMux) {
	return func(router *http.ServeMux) {
		router.HandleFunc("/__dev-server/ws-live-reload.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			fmt.Fprint(w, script)
		})
	}
}
