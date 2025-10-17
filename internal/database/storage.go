package database

import "integrity-monitor/pkg/models"

// Storage defines the interface for database operations
type Storage interface {
	SaveUtility(util *models.Utility) error
	GetUtility(path string) (*models.Utility, error)
	GetAllUtilities() ([]*models.Utility, error)
	SaveAlert(alert *models.Alert) error
	GetRecentAlerts(limit int) ([]*models.Alert, error)
	Close() error
}
