.PHONY: build install clean test init scan monitor

# Build the application
build:
	go build -o integrity-monitor cmd/integrity-monitor/main.go

# Install to system (requires root)
install: build
	mkdir -p /var/lib/integrity-monitor
	mkdir -p /etc/integrity-monitor
	cp integrity-monitor /usr/local/bin/
	chmod +x /usr/local/bin/integrity-monitor
	cp configs/config.yaml /etc/integrity-monitor/config.yaml
	@echo "Installation complete!"
	@echo "Run 'sudo integrity-monitor -init' to initialize the database"

# Clean build artifacts
clean:
	rm -f integrity-monitor

# Download dependencies
deps:
	go mod download
	go mod tidy

# Initialize database
init: build
	sudo ./integrity-monitor -init

# Run one-time scan
scan: build
	sudo ./integrity-monitor -scan

# Start monitoring
monitor: build
	sudo ./integrity-monitor

# Install systemd service
install-service:
	cp scripts/integrity-monitor.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable integrity-monitor
	@echo "Service installed. Start with: sudo systemctl start integrity-monitor"
