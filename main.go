package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"log/slog"

	"github.com/antelman107/net-wait-go/wait"
	"github.com/julienschmidt/httprouter"
	"github.com/lmittmann/tint"
	"github.com/pietjan/dev-server/proxy"
	"github.com/pietjan/dev-server/runner"
	"github.com/pietjan/dev-server/watcher"

	_ "embed"
)

const (
	configFile = `.dev-server.yml`
)

//go:embed live-reload.js
var script string

func main() {
	config := loadConfig()

	logger := slog.New(tint.NewHandler(os.Stdout, nil))
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	watcher := watcher.New(config.watcher()...)
	runner := runner.New(config.runner()...)
	proxy := proxy.New(
		proxy.Target(fmt.Sprintf(`http://localhost:%d`, config.Proxy)),
	)

	if !wait.New().Do(config.Wait) {
		logger.Error(`timeout waiting for services`)
		os.Exit(1)
	}

	messages := make(chan string)

	go func() {
		log := logger.With(`source`, `routine`)

		if err := runner.Exec(); err != nil {
			log.With(`error`, err).Error(`runner error`)
		}

		for {
			changes, err := watcher.Changes()
			if err != nil {
				log.With(`error`, err).Error(`watcher error`)
			}

			if len(changes) > 0 {
				logger.With(`changes`, changes).Info(`files changed`)

				if err := runner.Exec(); err != nil {
					log.With(`error`, err).Error(`runner error`)
				}

				time.Sleep(time.Millisecond * config.Interval)
				messages <- strings.Join(changes, `,`)
			}

			time.Sleep(time.Millisecond)
		}
	}()

	router := httprouter.New()
	router.GET(`/dev-server/sse`, sse(messages, logger))
	router.GET(`/dev-server/live-reload.js`, hotReloadScript)

	router.NotFound = proxy

	server := &http.Server{Addr: fmt.Sprintf(`:%d`, config.Server), Handler: router}

	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		if err := runner.Stop(); err != nil {
			logger.With(`error`, err).Error(`failed to stop target`)
		}
		if err := server.Shutdown(context.Background()); err != nil {
			logger.With(`error`, err).Error(`failed to stop server`)
		}
		close(done)
	}()

	if !wait.New().Do([]string{fmt.Sprintf(`:%d`, config.Proxy)}) {
		logger.Error(`target is not available`)
		os.Exit(1)
	}

	go open(fmt.Sprintf(`http://localhost:%d`, config.Server))
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	<-done
}

func hotReloadScript(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprint(w, script)
}

func sse(message chan string, logger *slog.Logger) httprouter.Handle {
	log := logger.With(`source`, `sse`)

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set(`Content-Type`, `text/event-stream`)
		w.Header().Set(`Cache-Control`, `no-cache`)
		w.Header().Set(`Connection`, `keep-alive`)

		log.Info(`client connected`)

		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Error(`faild to init http.Flusher`)
		}

		for {
			select {
			case m := <-message:
				fmt.Fprintf(w, "data: %s\n\n", m)
				flusher.Flush()
			case <-r.Context().Done():
				log.Info(`client disconnected`)
				return
			}
		}
	}
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case `windows`:
		cmd = `cmd`
		args = []string{`/c`, `start`}
	case `darwin`:
		cmd = `open`
	case `linux`:
		if _, err := os.Stat(`/mnt/c/WINDOWS/system32/wsl.exe`); err == nil {
			cmd = `explorer.exe`
			break
		}
		fallthrough
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = `xdg-open`
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
