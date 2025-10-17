package checksum

import (
	"fmt"
	"log"
	"os"
	"time"

	"integrity-monitor/internal/database"
	"integrity-monitor/pkg/models"
)

type Comparator struct {
	storage database.Storage
}

func NewComparator(storage database.Storage) *Comparator {
	return &Comparator{storage: storage}
}

// CheckFile verifies if a file's checksum matches the stored value
func (c *Comparator) CheckFile(filePath string) (*models.Alert, error) {
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Calculate current checksum
	currentChecksum, err := CalculateSHA256(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Get stored utility
	storedUtil, err := c.storage.GetUtility(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored utility: %w", err)
	}

	// If no stored checksum, this is a new file
	if storedUtil == nil {
		log.Printf("New utility detected: %s", filePath)
		return nil, nil
	}

	// Compare checksums
	if storedUtil.Checksum != currentChecksum {
		alert := &models.Alert{
			UtilityPath: filePath,
			OldChecksum: storedUtil.Checksum,
			NewChecksum: currentChecksum,
			DetectedAt:  time.Now(),
			Severity:    "critical",
		}

		// Save alert
		if err := c.storage.SaveAlert(alert); err != nil {
			log.Printf("Failed to save alert: %v", err)
		}

		return alert, nil
	}

	// Update last modified time if file changed but checksum is same
	if !fileInfo.ModTime().Equal(storedUtil.LastModified) {
		updatedUtil := &models.Utility{
			Path:         filePath,
			Checksum:     currentChecksum,
			LastModified: fileInfo.ModTime(),
			Size:         fileInfo.Size(),
		}
		c.storage.SaveUtility(updatedUtil)
	}

	return nil, nil
}

// StoreChecksum stores or updates a utility's checksum in the database
func (c *Comparator) StoreChecksum(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	checksum, err := CalculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	util := &models.Utility{
		Path:         filePath,
		Checksum:     checksum,
		LastModified: fileInfo.ModTime(),
		Size:         fileInfo.Size(),
	}

	return c.storage.SaveUtility(util)
}
