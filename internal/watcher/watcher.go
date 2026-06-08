package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher wraps fsnotify to provide debounced file change events.
type Watcher struct {
	watcher  *fsnotify.Watcher
	OnChange chan struct{}
}

// NewWatcher initializes a new file watcher for a given directory.
func NewWatcher(dir string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	onChange := make(chan struct{})

	go func() {
		var mu sync.Mutex
		var timer *time.Timer

		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				// Ignore non-go files and auto-generated wire code
				if !strings.HasSuffix(event.Name, ".go") || strings.Contains(event.Name, "gox_wire_gen.go") {
					continue
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					mu.Lock()
					if timer != nil {
						timer.Stop()
					}
					// Debounce for 500ms
					timer = time.AfterFunc(500*time.Millisecond, func() {
						onChange <- struct{}{}
					})
					mu.Unlock()
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden dirs (like .git) and bin folder
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "bin" {
				return filepath.SkipDir
			}
			return w.Add(path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher:  w,
		OnChange: onChange,
	}, nil
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}
