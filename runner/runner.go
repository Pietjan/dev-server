package runner

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"
)

type Option = func(*runner)

type Runner interface {
	Exec() error
	Stop() error
}

// New Runner
func New(options ...Option) Runner {
	r := &runner{}

	for _, fn := range options {
		fn(r)
	}

	return r
}

func Build(name string, args ...string) func(*runner) {
	return func(r *runner) {
		r.build = append(r.build, command{name: name, args: args})
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
	build  []command
	target command
	cmd    *exec.Cmd
}

// Stop implements Runner.
func (r *runner) Stop() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.cmd != nil {
		err := r.cmd.Process.Kill()
		if err != nil {
			return err
		}
	}

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

	if err := build(r.build); err != nil {
		return err
	}

	if r.cmd != nil {
		err := r.cmd.Process.Kill()
		if err != nil {
			return err
		}
	}

	r.cmd = toCommand(r.target)
	go func() {
		if err := r.cmd.Run(); err != nil {
			log.Println(err)
		}
	}()

	return nil
}

func build(commands []command) error {
	for _, c := range commands {
		cmd := toCommand(c)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func toCommand(c command) *exec.Cmd {
	cmd := exec.Command(c.name, c.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd
}
