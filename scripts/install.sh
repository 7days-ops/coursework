#!/bin/bash

# Integrity Monitor Installation Script
# This script installs and configures the integrity monitor

set -e

echo "=== Integrity Monitor Installation ==="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Error: This script must be run as root (use sudo)"
    exit 1
fi

# Create necessary directories
echo "Creating directories..."
mkdir -p /var/lib/integrity-monitor
mkdir -p /var/log
mkdir -p /etc/integrity-monitor

# Copy configuration
echo "Installing configuration..."
if [ -f "../configs/config.yaml" ]; then
    cp ../configs/config.yaml /etc/integrity-monitor/config.yaml
    echo "Configuration installed to /etc/integrity-monitor/config.yaml"
else
    echo "Warning: config.yaml not found, will use defaults"
fi

# Build the application
echo "Building application..."
cd ..
go build -o integrity-monitor cmd/integrity-monitor/main.go

# Install binary
echo "Installing binary..."
cp integrity-monitor /usr/local/bin/
chmod +x /usr/local/bin/integrity-monitor

echo ""
echo "=== Installation Complete ==="
echo ""
echo "Next steps:"
echo "1. Initialize the database with current system state:"
echo "   sudo integrity-monitor -init"
echo ""
echo "2. Start monitoring:"
echo "   sudo integrity-monitor"
echo ""
echo "Or create a systemd service for automatic startup."
