package watcher

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

type Option = func(*watcher)

type Watcher interface {
	Changes() ([]string, error)
}

// New Watcher
func New(options ...Option) Watcher {
	w := &watcher{
		watch: os.DirFS(`.`),
	}

	for _, fn := range options {
		fn(w)
	}

	return w
}

func ExcludeRegex(patterns ...string) func(*watcher) {
	return func(w *watcher) {
		for _, pattern := range patterns {
			w.excludeRegex = append(w.excludeRegex, regexp.MustCompile(pattern))
		}
	}
}

func ExcludePattern(patterns ...string) func(*watcher) {
	return func(w *watcher) {
		for _, pattern := range patterns {
			if !doublestar.ValidatePathPattern(pattern) {
				panic("invalid pattern " + pattern)
			}

			w.excludePattern = append(w.excludePattern, pattern)
		}
	}
}

type watcher struct {
	mutex          sync.Mutex
	watch          fs.FS
	lastCheck      time.Time
	excludeRegex   []*regexp.Regexp
	excludePattern []string
	watched        []file
}

type file struct {
	name string
	hash string
}

// Changes implements Watcher.
func (w *watcher) Changes() (changes []string, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	checkTime := time.Now()
	if w.lastCheck.IsZero() {
		w.lastCheck = checkTime
		return
	}

	root := "."

	err = fs.WalkDir(w.watch, root, func(pathStr string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		relPath, _ := filepath.Rel(root, pathStr)

		if isExclude(relPath, w.excludePattern, w.excludeRegex) {
			return nil
		}

		f, err := d.Info()
		if err != nil {
			return err
		}

		if f.ModTime().Before(w.lastCheck) {
			return nil
		}

		b, err := fs.ReadFile(w.watch, pathStr)
		if err != nil {
			return err
		}

		sum := md5.Sum(bytes.TrimSpace(b))
		hash := hex.EncodeToString(sum[:])
		index := getWatched(pathStr, w.watched)

		if index < 0 {
			slog.Debug("file-added", "path", relPath, "hash", hash)
			w.watched = append(w.watched, file{
				name: pathStr,
				hash: hash,
			})
		} else {
			if w.watched[index].hash == hash {
				return nil
			}

			w.watched[index].hash = hash
		}

		changes = append(changes, pathStr)

		return nil
	})

	if len(changes) > 0 && err == nil {
		w.lastCheck = checkTime
	}

	return
}

func getWatched(name string, watched []file) int {
	for i, f := range watched {
		if f.name == name {
			return i
		}
	}

	return -1
}

func isExclude(file string, pattern []string, regexes []*regexp.Regexp) bool {
	for _, p := range pattern {
		if matched, _ := doublestar.PathMatch(p, file); matched {
			slog.Debug("file-excluded", "path", file, "pattern", p)
			return true
		}
	}

	for _, r := range regexes {
		if r.MatchString(file) {
			slog.Debug("file-excluded", "path", file, "regex", r.String())
			return true
		}
	}

	return false
}
