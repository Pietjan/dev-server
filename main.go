package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/antelman107/net-wait-go/wait"
	"github.com/pietjan/dev-server/app"
	"github.com/pietjan/dev-server/pkg/config"
	"github.com/pietjan/dev-server/pkg/logger"
)

func main() {
	config, err := config.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(config.Build.Command) == 0 {
		fmt.Println("build command is empty (--build.cmd)")
		os.Exit(1)
	}

	if len(config.Build.Bin) == 0 {
		fmt.Println("binary path is empty (--build.bin)")
		os.Exit(1)
	}

	slog.SetDefault(logger.New(config.Debug))

	// wait for dependant services
	if !wait.New().Do(config.Wait.For) {
		fmt.Println("timeout waiting for services")
		os.Exit(1)
	}

	app := app.New(config)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		signal := <-signals
		slog.Debug("cleanup", "signal", signal)
		if err := app.Stop(); err != nil {
			slog.Error("cleanup", "error", err)
		}
	}()

	fmt.Printf("\n\033[36m  dev-server: http://localhost:%d\033[0m\n", config.Proxy.Port)

	if err := app.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("app", "error", err)
	}
}
