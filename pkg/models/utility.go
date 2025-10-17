package models

import "time"

// Utility represents a system utility file
type Utility struct {
	ID           int64     `json:"id"`
	Path         string    `json:"path"`
	Checksum     string    `json:"checksum"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Alert represents a security alert for a modified utility
type Alert struct {
	UtilityPath    string    `json:"utility_path"`
	OldChecksum    string    `json:"old_checksum"`
	NewChecksum    string    `json:"new_checksum"`
	DetectedAt     time.Time `json:"detected_at"`
	Severity       string    `json:"severity"` // critical, high, medium
}
