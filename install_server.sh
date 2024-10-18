#!/bin/bash

# Ensure the script is executed from the correct directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || { echo "Failed to change directory to $SCRIPT_DIR"; exit 1; }

# Variables
SERVICE_FILE="$SCRIPT_DIR/go-server.service"
SYSTEMD_DIR="/etc/systemd/system"
SERVICE_NAME="go-server.service"
GO_EXECUTABLE="buildHandler"

# Build the Go executable
echo "Building Go executable..."
go build -o "$GO_EXECUTABLE" .

# Copy the systemd service file to the systemd directory
echo "Copying systemd service file..."
sudo cp "$SERVICE_FILE" "$SYSTEMD_DIR"

# Reload systemd to recognize the new service
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

# Enable the service to start on boot
echo "Enabling $SERVICE_NAME..."
sudo systemctl enable "$SERVICE_NAME"

# Start the service
echo "Starting $SERVICE_NAME..."
sudo systemctl start "$SERVICE_NAME"

# Check the status of the service
echo "Checking the status of $SERVICE_NAME..."
sudo systemctl status "$SERVICE_NAME"

echo "Setup completed successfully."
exit 0
