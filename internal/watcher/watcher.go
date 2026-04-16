// Package watcher wraps fsnotify to detect new/updated CSV files in a folder,
// debounces rapid-fire events, and filters out files the app itself just wrote.
package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Processor func(path string)

const (
	debouncePeriod  = 500 * time.Millisecond
	selfWriteIgnore = 2 * time.Second
)

type Watcher struct {
	proc Processor

	mu      sync.Mutex
	folder  string
	fsw     *fsnotify.Watcher
	timers  map[string]*time.Timer
	recent  map[string]time.Time
	paused  bool
	stopCh  chan struct{}
	running bool
}

func New(folder string, proc Processor) *Watcher {
	return &Watcher{
		proc:   proc,
		folder: folder,
		timers: make(map[string]*time.Timer),
		recent: make(map[string]time.Time),
	}
}

func (w *Watcher) Folder() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.folder
}

func (w *Watcher) Start() error {
	w.mu.Lock()
	if w.running || w.folder == "" {
		w.mu.Unlock()
		return nil
	}
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		w.mu.Unlock()
		return err
	}
	if err := fsw.Add(w.folder); err != nil {
		fsw.Close()
		w.mu.Unlock()
		return err
	}
	w.fsw = fsw
	w.stopCh = make(chan struct{})
	w.running = true
	folder := w.folder
	w.mu.Unlock()

	go w.loop()
	go w.initialSweep(folder)
	return nil
}

func (w *Watcher) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	close(w.stopCh)
	if w.fsw != nil {
		w.fsw.Close()
		w.fsw = nil
	}
	for _, t := range w.timers {
		t.Stop()
	}
	w.timers = make(map[string]*time.Timer)
	w.mu.Unlock()
}

func (w *Watcher) SetFolder(folder string) error {
	w.Stop()
	w.mu.Lock()
	w.folder = folder
	w.mu.Unlock()
	return w.Start()
}

func (w *Watcher) SetPaused(p bool) {
	w.mu.Lock()
	w.paused = p
	w.mu.Unlock()
}

func (w *Watcher) IsPaused() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.paused
}

// MarkSelfWritten marks a file path as written by the app itself so the
// ensuing fsnotify event is not processed as a new input.
func (w *Watcher) MarkSelfWritten(path string) {
	w.mu.Lock()
	w.recent[filepath.Clean(path)] = time.Now()
	w.mu.Unlock()
}

func (w *Watcher) loop() {
	for {
		w.mu.Lock()
		fsw := w.fsw
		stopCh := w.stopCh
		w.mu.Unlock()
		if fsw == nil {
			return
		}
		select {
		case <-stopCh:
			return
		case ev, ok := <-fsw.Events:
			if !ok {
				return
			}
			w.handleEvent(ev)
		case _, ok := <-fsw.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *Watcher) handleEvent(ev fsnotify.Event) {
	if ev.Op&(fsnotify.Create|fsnotify.Write) == 0 {
		return
	}
	if !isCSV(ev.Name) {
		return
	}
	if filepath.Dir(ev.Name) != w.Folder() {
		return
	}
	if w.isRecentlyWritten(ev.Name) {
		return
	}
	w.scheduleProcess(ev.Name)
}

func (w *Watcher) isRecentlyWritten(path string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	key := filepath.Clean(path)
	t, ok := w.recent[key]
	if !ok {
		return false
	}
	if time.Since(t) > selfWriteIgnore {
		delete(w.recent, key)
		return false
	}
	return true
}

func (w *Watcher) scheduleProcess(path string) {
	w.mu.Lock()
	if timer, exists := w.timers[path]; exists {
		timer.Reset(debouncePeriod)
		w.mu.Unlock()
		return
	}
	timer := time.AfterFunc(debouncePeriod, func() {
		w.mu.Lock()
		delete(w.timers, path)
		paused := w.paused
		w.mu.Unlock()
		if paused {
			return
		}
		if w.isRecentlyWritten(path) {
			return
		}
		info, err := os.Stat(path)
		if err != nil || info.IsDir() || info.Size() == 0 {
			return
		}
		w.proc(path)
	})
	w.timers[path] = timer
	w.mu.Unlock()
}

func (w *Watcher) initialSweep(folder string) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !isCSV(e.Name()) {
			continue
		}
		if w.IsPaused() {
			return
		}
		w.proc(filepath.Join(folder, e.Name()))
	}
}

func isCSV(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".csv")
}
