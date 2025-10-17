package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database       DatabaseConfig `yaml:"database"`
	MonitoredPaths []string       `yaml:"monitored_paths"`
	ScanInterval   int            `yaml:"scan_interval"` // seconds
	EnableWatcher  bool           `yaml:"enable_watcher"`
	LogFile        string         `yaml:"log_file"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func Default() *Config {
	return &Config{
		Database: DatabaseConfig{
			Path: "/var/lib/integrity-monitor/checksums.db",
		},
		MonitoredPaths: []string{
			"/bin",
			"/sbin",
			"/usr/bin",
			"/usr/sbin",
			"/usr/local/bin",
			"/usr/local/sbin",
		},
		ScanInterval:  300, // 5 minutes
		EnableWatcher: true,
		LogFile:       "/var/log/integrity-monitor.log",
	}
}
