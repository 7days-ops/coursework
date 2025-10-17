package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"integrity-monitor/internal/checksum"
	"integrity-monitor/internal/config"
	"integrity-monitor/internal/database"
	"integrity-monitor/internal/notifier"
	"integrity-monitor/internal/scanner"
	"integrity-monitor/internal/watcher"
)

const version = "1.0.0"

func main() {
	// Define command line flags
	initCmd := flag.Bool("init", false, "Initialize database with current system state")
	scanCmd := flag.Bool("scan", false, "Perform a one-time scan of all utilities")
	configPath := flag.String("config", "/etc/integrity-monitor/config.yaml", "Path to configuration file")
	versionFlag := flag.Bool("version", false, "Show version information")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("Integrity Monitor v%s\n", version)
		return
	}

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	storage, err := database.NewSQLiteStorage(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer storage.Close()

	comp := checksum.NewComparator(storage)
	scan := scanner.NewScanner(cfg.MonitoredPaths)

	// Handle commands
	switch {
	case *initCmd:
		initializeDatabase(scan, comp)
	case *scanCmd:
		performScan(scan, comp, cfg)
	default:
		// Default to monitoring
		startMonitoring(cfg, storage, scan, comp)
	}
}

func loadConfig(configPath string) (*config.Config, error) {
	// Try to load from file
	if _, err := os.Stat(configPath); err == nil {
		return config.Load(configPath)
	}

	// Use default configuration
	log.Printf("Config file not found, using defaults")
	return config.Default(), nil
}

func initializeDatabase(scan *scanner.Scanner, comp *checksum.Comparator) {
	log.Println("Initializing database with current system state...")

	utilities, err := scan.ScanAll()
	if err != nil {
		log.Fatalf("Failed to scan utilities: %v", err)
	}

	log.Printf("Found %d utilities to process", len(utilities))

	successCount := 0
	for i, util := range utilities {
		if (i+1)%100 == 0 {
			log.Printf("Progress: %d/%d utilities processed", i+1, len(utilities))
		}

		if err := comp.StoreChecksum(util); err != nil {
			log.Printf("Warning: failed to store checksum for %s: %v", util, err)
			continue
		}
		successCount++
	}

	log.Printf("Initialization complete! Stored checksums for %d/%d utilities", successCount, len(utilities))
}

func performScan(scan *scanner.Scanner, comp *checksum.Comparator, cfg *config.Config) {
	log.Println("Performing one-time scan...")

	utilities, err := scan.ScanAll()
	if err != nil {
		log.Fatalf("Failed to scan utilities: %v", err)
	}

	log.Printf("Checking %d utilities", len(utilities))

	notif := notifier.NewTTYNotifier(cfg.LogFile)
	alertCount := 0

	for _, util := range utilities {
		alert, err := comp.CheckFile(util)
		if err != nil {
			log.Printf("Error checking %s: %v", util, err)
			continue
		}

		if alert != nil {
			alertCount++
			log.Printf("ALERT: %s has been modified!", util)
			notif.SendAlert(alert)
		}
	}

	if alertCount == 0 {
		log.Println("Scan complete: No modifications detected")
	} else {
		log.Printf("Scan complete: %d modified utilities detected!", alertCount)
	}
}

func startMonitoring(cfg *config.Config, storage database.Storage, scan *scanner.Scanner, comp *checksum.Comparator) {
	log.Println("Starting Integrity Monitor...")
	log.Printf("Monitoring paths: %v", cfg.MonitoredPaths)
	log.Printf("Scan interval: %d seconds", cfg.ScanInterval)

	notif := notifier.NewTTYNotifier(cfg.LogFile)

	// Start periodic scanner
	go startPeriodicScan(scan, comp, notif, cfg.ScanInterval)

	// Start file watcher if enabled
	if cfg.EnableWatcher {
		handler := watcher.CreateFileChangeHandler(comp, notif)
		w, err := watcher.NewWatcher(cfg.MonitoredPaths, handler)
		if err != nil {
			log.Fatalf("Failed to create watcher: %v", err)
		}
		defer w.Close()

		go func() {
			if err := w.Start(); err != nil {
				log.Fatalf("Watcher error: %v", err)
			}
		}()
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Monitoring active. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("Shutting down...")
}

func startPeriodicScan(scan *scanner.Scanner, comp *checksum.Comparator, notif notifier.Notifier, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Starting periodic scan...")

		utilities, err := scan.ScanAll()
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		alertCount := 0
		for _, util := range utilities {
			alert, err := comp.CheckFile(util)
			if err != nil {
				log.Printf("Error checking %s: %v", util, err)
				continue
			}

			if alert != nil {
				alertCount++
				notif.SendAlert(alert)
			}
		}

		if alertCount > 0 {
			log.Printf("Periodic scan complete: %d alerts generated", alertCount)
		} else {
			log.Println("Periodic scan complete: No modifications detected")
		}
	}
}
