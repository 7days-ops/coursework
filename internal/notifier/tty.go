package notifier

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"integrity-monitor/pkg/models"
)

type TTYNotifier struct {
	logFile string
}

func NewTTYNotifier(logFile string) *TTYNotifier {
	return &TTYNotifier{logFile: logFile}
}

func (n *TTYNotifier) SendAlert(alert *models.Alert) error {
	message := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════╗
║              ⚠️  SECURITY ALERT - UTILITY MODIFIED  ⚠️        ║
╠══════════════════════════════════════════════════════════════╣
║ Path:          %s
║ Severity:      %s
║ Old Checksum:  %s
║ New Checksum:  %s
║ Detected At:   %s
║
║ WARNING: A system utility has been modified!
║ This could indicate a security breach or malicious activity.
║ Please investigate immediately!
╚══════════════════════════════════════════════════════════════╝
`,
		alert.UtilityPath,
		strings.ToUpper(alert.Severity),
		alert.OldChecksum[:16]+"...",
		alert.NewChecksum[:16]+"...",
		alert.DetectedAt.Format("2006-01-02 15:04:05"),
	)

	// Log to file
	if err := n.logToFile(message); err != nil {
		log.Printf("Failed to log alert to file: %v", err)
	}

	// Send to all active TTYs
	if err := n.broadcastToTTYs(message); err != nil {
		log.Printf("Failed to broadcast to TTYs: %v", err)
	}

	return nil
}

func (n *TTYNotifier) logToFile(message string) error {
	f, err := os.OpenFile(n.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(message + "\n")
	return err
}

func (n *TTYNotifier) broadcastToTTYs(message string) error {
	// Find all TTY and PTS devices
	ttys := []string{}

	// Check /dev/tty*
	ttyMatches, _ := filepath.Glob("/dev/tty[0-9]*")
	ttys = append(ttys, ttyMatches...)

	// Check /dev/pts/*
	ptsDir := "/dev/pts"
	if entries, err := os.ReadDir(ptsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && entry.Name() != "ptmx" {
				ttys = append(ttys, filepath.Join(ptsDir, entry.Name()))
			}
		}
	}

	// Try to write to each TTY
	successCount := 0
	for _, tty := range ttys {
		if err := n.writeToTTY(tty, message); err == nil {
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf("failed to write to any TTY devices")
	}

	log.Printf("Alert broadcast to %d TTY devices", successCount)
	return nil
}

func (n *TTYNotifier) writeToTTY(ttyPath string, message string) error {
	// Check if we have write permission
	info, err := os.Stat(ttyPath)
	if err != nil {
		return err
	}

	mode := info.Mode()
	// Check if we have write permission (owner, group, or others)
	if mode&0222 == 0 {
		return fs.ErrPermission
	}

	// Try to write to the TTY
	f, err := os.OpenFile(ttyPath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(message + "\n")
	return err
}
