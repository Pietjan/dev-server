package watcher

import (
	"io/fs"
	"os"
	"regexp"
	"sync"
	"time"
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

func Exclude(pattern string) func(*watcher) {
	return func(w *watcher) {
		w.exclude = append(w.exclude, regexp.MustCompile(pattern))
	}
}

type watcher struct {
	mutex     sync.Mutex
	watch     fs.FS
	lastCheck time.Time
	exclude   []*regexp.Regexp
}

// Changes implements Watcher.
func (w *watcher) Changes() (changes []string, err error) {
	w.mutex.Lock()
	checkTime := time.Now()
	defer w.mutex.Unlock()

	if w.lastCheck.IsZero() {
		w.lastCheck = checkTime
		return
	}

	err = fs.WalkDir(w.watch, `.`, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		for _, e := range w.exclude {
			if e.MatchString(path) {
				return nil
			}
		}

		f, err := d.Info()
		if err != nil {
			return err
		}

		if f.ModTime().After(w.lastCheck) {
			changes = append(changes, path)
		}

		return nil
	})

	if len(changes) > 0 && err == nil {
		w.lastCheck = checkTime
	}

	return
}
