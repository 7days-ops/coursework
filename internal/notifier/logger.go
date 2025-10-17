package notifier

import (
	"fmt"
	"os"
	"time"

	"integrity-monitor/pkg/models"
)

type FileLogger struct {
	logFile string
}

func NewFileLogger(logFile string) *FileLogger {
	return &FileLogger{logFile: logFile}
}

func (l *FileLogger) SendAlert(alert *models.Alert) error {
	message := fmt.Sprintf("[%s] ALERT: %s - Utility %s modified (old: %s, new: %s)\n",
		time.Now().Format("2006-01-02 15:04:05"),
		alert.Severity,
		alert.UtilityPath,
		alert.OldChecksum,
		alert.NewChecksum,
	)

	f, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(message)
	return err
}
