package notifier

import "integrity-monitor/pkg/models"

// Notifier defines the interface for sending alerts
type Notifier interface {
	SendAlert(alert *models.Alert) error
}
