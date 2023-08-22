package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pietjan/dev-server/runner"
	"github.com/pietjan/dev-server/watcher"
)

type config struct {
	Interval time.Duration `json:"interval"`
	Exclude  []string      `json:"exclude"`
	Build    string        `json:"build"`
	Target   string        `json:"target"`
	Wait     []string      `json:"wait"`
	Server   int           `json:"server"`
	Proxy    int           `json:"proxy"`
}

func (c config) watcher() (options []watcher.Option) {
	for _, exclude := range c.Exclude {
		options = append(options, watcher.Exclude(exclude))
	}
	return
}

func (c config) runner() (options []runner.Option) {
	options = append(options, runner.Build(`go`, `build`, `-o`, c.Target, c.Build))
	options = append(options, runner.Target(c.Target))

	return
}

func loadConfig() config {
	c := config{}
	fromFile(&c)
	defaults(&c)
	return c
}

func fromFile(c *config) error {
	for _, fileName := range []string{`.dev-server.json`, `.dev-server.yml`, `.dev-server.yaml`} {
		if _, err := os.Stat(fileName); err == nil {
			data, err := os.ReadFile(fileName)
			if err != nil {
				continue
			}

			if strings.HasSuffix(fileName, `.json`) {
				return json.Unmarshal(data, c)
			} else {
				return yaml.Unmarshal(data, c)
			}
		}

	}

	return nil
}

func defaults(c *config) {
	if c.Interval == 0 {
		c.Interval = 300
	}

	if len(c.Exclude) == 0 {
		c.Exclude = []string{`^bin/`, `^\.`, `\/\.(\w+)$`}
	}

	if len(c.Build) == 0 {
		c.Build = `./`
	}

	if len(c.Target) == 0 {
		c.Target = `./bin/__dev-server_target`
	}

	if c.Server == 0 {
		c.Server = 42069
	}

	if c.Proxy == 0 {
		c.Proxy = 8080
	}
}
