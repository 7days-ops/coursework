package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"integrity-monitor/pkg/models"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

func (s *SQLiteStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS utilities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE,
		checksum TEXT NOT NULL,
		last_modified DATETIME NOT NULL,
		size INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_utilities_path ON utilities(path);
	CREATE INDEX IF NOT EXISTS idx_utilities_checksum ON utilities(checksum);

	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		utility_path TEXT NOT NULL,
		old_checksum TEXT NOT NULL,
		new_checksum TEXT NOT NULL,
		detected_at DATETIME NOT NULL,
		severity TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_alerts_detected_at ON alerts(detected_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStorage) SaveUtility(util *models.Utility) error {
	query := `
	INSERT INTO utilities (path, checksum, last_modified, size, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		checksum = excluded.checksum,
		last_modified = excluded.last_modified,
		size = excluded.size,
		updated_at = excluded.updated_at
	`

	now := time.Now()
	_, err := s.db.Exec(query, util.Path, util.Checksum, util.LastModified, util.Size, now, now)
	return err
}

func (s *SQLiteStorage) GetUtility(path string) (*models.Utility, error) {
	query := `SELECT id, path, checksum, last_modified, size, created_at, updated_at
	          FROM utilities WHERE path = ?`

	var util models.Utility
	err := s.db.QueryRow(query, path).Scan(
		&util.ID, &util.Path, &util.Checksum, &util.LastModified,
		&util.Size, &util.CreatedAt, &util.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &util, nil
}

func (s *SQLiteStorage) GetAllUtilities() ([]*models.Utility, error) {
	query := `SELECT id, path, checksum, last_modified, size, created_at, updated_at
	          FROM utilities ORDER BY path`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var utilities []*models.Utility
	for rows.Next() {
		var util models.Utility
		if err := rows.Scan(
			&util.ID, &util.Path, &util.Checksum, &util.LastModified,
			&util.Size, &util.CreatedAt, &util.UpdatedAt,
		); err != nil {
			return nil, err
		}
		utilities = append(utilities, &util)
	}

	return utilities, rows.Err()
}

func (s *SQLiteStorage) SaveAlert(alert *models.Alert) error {
	query := `INSERT INTO alerts (utility_path, old_checksum, new_checksum, detected_at, severity)
	          VALUES (?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, alert.UtilityPath, alert.OldChecksum, alert.NewChecksum,
		alert.DetectedAt, alert.Severity)
	return err
}

func (s *SQLiteStorage) GetRecentAlerts(limit int) ([]*models.Alert, error) {
	query := `SELECT utility_path, old_checksum, new_checksum, detected_at, severity
	          FROM alerts ORDER BY detected_at DESC LIMIT ?`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		var alert models.Alert
		if err := rows.Scan(&alert.UtilityPath, &alert.OldChecksum, &alert.NewChecksum,
			&alert.DetectedAt, &alert.Severity); err != nil {
			return nil, err
		}
		alerts = append(alerts, &alert)
	}

	return alerts, rows.Err()
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
