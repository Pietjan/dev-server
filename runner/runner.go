package runner

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"

	"log/slog"
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

func Build(name string, args ...string) func(*runner) {
	return func(r *runner) {
		r.build = command{name: name, args: args}
	}
}

func Target(name string, args ...string) func(*runner) {
	return func(r *runner) {
		r.target = command{name: name, args: args}
	}
}

type runner struct {
	mutex  sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	log    *slog.Logger
	build  command
	target command
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

	build := toExecCmd(r.ctx, r.build)
	if err := build.Run(); err != nil {
		return err
	}

	if r.cmd != nil && r.cmd.Process != nil {
		if err := r.cmd.Process.Kill(); err != nil {
			return err
		}
	}

	r.cmd = toExecCmd(r.ctx, r.target)
	if err := r.cmd.Start(); err != nil {
		log.Println(err)
	}

	return nil
}

func toExecCmd(ctx context.Context, c command) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.name, c.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	return cmd
}
