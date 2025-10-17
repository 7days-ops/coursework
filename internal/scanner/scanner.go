package scanner

import (
	"fmt"
	"os"
	"path/filepath"
)

type Scanner struct {
	paths []string
}

func NewScanner(paths []string) *Scanner {
	return &Scanner{paths: paths}
}

// ScanAll returns all executable files in monitored directories
func (s *Scanner) ScanAll() ([]string, error) {
	var utilities []string
	seen := make(map[string]bool)

	for _, path := range s.paths {
		files, err := s.scanDirectory(path)
		if err != nil {
			// Log error but continue with other directories
			fmt.Printf("Warning: failed to scan %s: %v\n", path, err)
			continue
		}

		for _, file := range files {
			if !seen[file] {
				utilities = append(utilities, file)
				seen[file] = true
			}
		}
	}

	return utilities, nil
}

func (s *Scanner) scanDirectory(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is executable
		if isExecutable(info) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	return files, nil
}

// isExecutable checks if a file has executable permissions
func isExecutable(info os.FileInfo) bool {
	mode := info.Mode()
	// Check if any execute bit is set (owner, group, or others)
	return mode&0111 != 0
}
