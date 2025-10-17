package watcher

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsWatcher     *fsnotify.Watcher
	paths         []string
	eventHandler  EventHandler
}

type EventHandler func(path string, event fsnotify.Op) error

func NewWatcher(paths []string, handler EventHandler) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		fsWatcher:    fsWatcher,
		paths:        paths,
		eventHandler: handler,
	}

	// Add all paths to watch
	for _, path := range paths {
		if err := fsWatcher.Add(path); err != nil {
			log.Printf("Warning: failed to watch %s: %v", path, err)
		} else {
			log.Printf("Watching directory: %s", path)
		}
	}

	return w, nil
}

func (w *Watcher) Start() error {
	log.Println("Starting file watcher...")

	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return fmt.Errorf("watcher events channel closed")
			}

			// Only handle write and create events
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {
				
				// Get absolute path
				absPath, err := filepath.Abs(event.Name)
				if err != nil {
					log.Printf("Failed to get absolute path for %s: %v", event.Name, err)
					continue
				}

				// Call event handler
				if err := w.eventHandler(absPath, event.Op); err != nil {
					log.Printf("Error handling event for %s: %v", absPath, err)
				}
			}

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}
