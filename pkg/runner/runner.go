package runner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Option = func(*runner)

type Runner interface {
	Exec() error
	Stop() error
}

// New Runner
func New(options ...Option) Runner {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	r := &runner{
		ctx:    ctx,
		cancel: cancel,
	}

	for _, fn := range options {
		fn(r)
	}

	return r
}

func Build(cmd string) func(*runner) {
	c := strings.Split(cmd, " ")
	return func(r *runner) {
		r.build = command{name: c[0], args: c[1:]}
	}
}

func Target(cmd string) func(*runner) {
	c := strings.Split(cmd, " ")
	return func(r *runner) {
		r.target = command{name: c[0], args: c[1:]}
	}
}

func Port(port int) func(*runner) {
	return func(r *runner) {
		r.port = port
	}
}

type runner struct {
	mutex  sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	build  command
	target command
	port   int
	cmd    *exec.Cmd
}

// Stop implements Runner.
func (r *runner) Stop() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.cancel()

	<-r.ctx.Done()

	return os.Remove(r.target.name)
}

type command struct {
	name string
	args []string
}

// Exec implements Runner.
func (r *runner) Exec() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	slog.Debug("runner-exec", "build", r.build.name, "target", r.target.name, "port", r.port)

	build := toExecCmd(r.ctx, r.build)

	if err := build.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if r.cmd != nil && r.cmd.Process != nil {
		if err := r.cmd.Process.Kill(); err != nil {
			return err
		}
	}

	slog.Debug("runner", "wait-for-port-free", r.port)
	ctx, cancel := context.WithTimeout(r.ctx, time.Second*3)
	if err := waitForPortFree(ctx, "localhost", r.port); err != nil {
		cancel()
		return err
	}
	cancel()
	slog.Debug("runner", "port-free", r.port)

	slog.Debug("runner", "start", r.target.name)

	r.cmd = toExecCmd(r.ctx, r.target)
	return r.cmd.Start()
}

func toExecCmd(ctx context.Context, c command) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.name, c.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	return cmd
}

func waitForPortFree(ctx context.Context, host string, port int) error {
	address := fmt.Sprintf("%s:%d", host, port)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout reached, port %d is still in use", port)
		default:
			ln, err := net.Listen("tcp", address)
			if err != nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			ln.Close()
			return nil
		}
	}
}
