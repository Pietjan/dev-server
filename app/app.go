package app

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/pietjan/dev-server/pkg/config"
	"github.com/pietjan/dev-server/pkg/runner"
	"github.com/pietjan/dev-server/pkg/server"
	"github.com/pietjan/dev-server/pkg/server/router"
	"github.com/pietjan/dev-server/pkg/watcher"
	"github.com/pietjan/events"

	_ "embed"
)

//go:embed assets/ws-live-reload.js
var scriptWS string

func New(config config.Settings) *App {
	app := App{
		event: events.NewProcessor(4, 100),
		runner: runner.New(
			runner.Build(config.Build.Command),
			runner.Target(config.Build.Bin),
			runner.Port(config.Proxy.Target),
		),
		watcher: watcher.New(
			watcher.ExcludePattern(config.Watcher.ExcludePattern...),
			watcher.ExcludeRegex(config.Watcher.ExcludeRegex...),
		),
		ticker: time.NewTicker(config.Watcher.Interval),
	}

	app.server = server.New(
		server.Port(config.Proxy.Port),
		server.Router(
			router.ProxyTarget(config.Proxy.Target),
			router.LiveReloadWS(scriptWS),
			router.WS(app.event),
		),
	)

	return &app
}

type App struct {
	event   *events.Processor
	runner  runner.Runner
	watcher watcher.Watcher
	server  *http.Server
	ticker  *time.Ticker
}

func (app *App) Start() error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
			}
		}()

		fmt.Println()
		if err := app.runner.Exec(); err != nil {
			log.Println(err)
		}
		fmt.Println()

		for {
			<-app.ticker.C
			changes, err := app.watcher.Changes()
			if err != nil {
				log.Println(err)
				continue
			}

			if len(changes) > 0 {
				for _, change := range changes {
					slog.Info("changed", "file", change)
				}

				fmt.Println()
				if err := app.runner.Exec(); err != nil {
					slog.Error("runner-exec", "error", err)
				}
				fmt.Println()

				app.event.Emit("message", "reload")
			}
		}
	}()

	return app.server.ListenAndServe()
}

func (app *App) Stop() error {
	app.ticker.Stop()
	app.event.Stop()
	app.server.Close()

	return app.runner.Stop()
}

func (app *App) On(event string, handler func(event *events.Event)) func() {
	return app.event.On(event, handler)
}
