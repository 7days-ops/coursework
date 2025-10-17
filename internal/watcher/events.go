package watcher

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"integrity-monitor/internal/checksum"
	"integrity-monitor/internal/notifier"
)

// CreateFileChangeHandler creates an event handler for file modifications
func CreateFileChangeHandler(comp *checksum.Comparator, notif notifier.Notifier) EventHandler {
	return func(path string, event fsnotify.Op) error {
		// Check if file is executable
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		// Only process executable files
		if info.Mode()&0111 == 0 {
			return nil
		}

		log.Printf("Detected change in %s (event: %s)", path, event.String())

		// Check the file's integrity
		alert, err := comp.CheckFile(path)
		if err != nil {
			return err
		}

		// If there's an alert, notify users
		if alert != nil {
			log.Printf("ALERT: Utility %s has been modified!", path)
			return notif.SendAlert(alert)
		}

		return nil
	}
}
