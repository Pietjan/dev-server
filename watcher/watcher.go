package watcher

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
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
	watched   []file
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

		if f.ModTime().Before(w.lastCheck) {
			return nil
		}

		b, err := fs.ReadFile(w.watch, path)
		if err != nil {
			return err
		}

		sum := md5.Sum(bytes.TrimSpace(b))
		hash := hex.EncodeToString(sum[:])
		index := getWatched(path, w.watched)

		if index < 0 {
			w.watched = append(w.watched, file{
				name: path,
				hash: hash,
			})
		} else {
			if w.watched[index].hash == hash {
				return nil
			}

			w.watched[index].hash = hash
		}

		changes = append(changes, path)

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
