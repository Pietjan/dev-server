package config

import (
	"encoding/json"
	"flag"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Settings struct {
	Build   Build   `json:"build"`
	Watcher Watcher `json:"watcher"`
	Proxy   Proxy   `json:"proxy"`
	Wait    Wait    `json:"wait"`
	Debug   bool    `json:"debug"`
}

type Build struct {
	Command string `json:"cmd"`
	Bin     string `json:"bin"`
}

type Watcher struct {
	Interval time.Duration `json:"interval"`
	Exclude  slice         `json:"exclude"`
}

type Proxy struct {
	Port   int `json:"port"`
	Target int `json:"target"`
}

type Wait struct {
	For slice `json:"for"`
}

type slice []string

func (s *slice) String() string {
	return strings.Join(*s, ", ")
}

func (s *slice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func Load() (Settings, error) {
	settings := Settings{
		Watcher: Watcher{
			Interval: time.Millisecond * 300,
		},
		Proxy: Proxy{
			Port:   42069,
			Target: 8080,
		},
	}
	if err := fromFile(&settings); err != nil {
		return settings, err
	}

	flag.StringVar(&settings.Build.Command, "build.cmd", settings.Build.Command, "build command")
	flag.StringVar(&settings.Build.Bin, "build.bin", settings.Build.Bin, "binary path")

	flag.DurationVar(&settings.Watcher.Interval, "watcher.interval", settings.Watcher.Interval, "watcher interval")
	flag.Var(&settings.Watcher.Exclude, "watcher.exclude", "exclude patterns")

	flag.IntVar(&settings.Proxy.Port, "proxy.port", settings.Proxy.Port, "proxy port")
	flag.IntVar(&settings.Proxy.Target, "proxy.target", settings.Proxy.Target, "proxy target")

	flag.Var(&settings.Wait.For, "wait.for", "wait for services")
	flag.BoolVar(&settings.Debug, "debug", settings.Debug, "debug mode")

	flag.Parse()

	return settings, nil
}

func fromFile(s *Settings) error {
	for _, fileName := range []string{".dev-server.json", ".dev-server.yml", ".dev-server.yaml"} {
		if _, err := os.Stat(fileName); err == nil {
			data, err := os.ReadFile(fileName)
			if err != nil {
				continue
			}

			if strings.HasSuffix(fileName, ".json") {
				return json.Unmarshal(data, s)
			} else {
				return yaml.Unmarshal(data, s)
			}
		}

	}

	return nil
}
