#!/bin/bash

# Load environment variables from a .env file
source ./.env

# Ensure the script is executed from the correct directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || { echo "Failed to change directory to $SCRIPT_DIR"; exit 1; }

# Prompt for User and Group
read -p "Enter the user to run the service: " SERVICE_USER
read -p "Enter the group to run the service: " SERVICE_GROUP

# Variables
SERVICE_FILE_TEMPLATE="$SCRIPT_DIR/go-server.service.template"
SERVICE_FILE="$SCRIPT_DIR/go-server.service"
SYSTEMD_DIR="/etc/systemd/system"
SERVICE_NAME="go-server.service"
GO_EXECUTABLE="buildHandler"

# Ensure Go is in the PATH
export PATH=$PATH:/usr/local/go/bin  # Adjust this path if Go is installed elsewhere

# Build the Go executable
echo "Building Go executable..."
go build -o "$GO_EXECUTABLE" .

# Generate the service file from the template
echo "Generating systemd service file..."
sed "s|{{USER}}|$SERVICE_USER|g; s|{{GROUP}}|$SERVICE_GROUP|g; s|{{WORKING_DIRECTORY}}|$SCRIPT_DIR|g; s|{{EXEC_START}}|$SCRIPT_DIR/$GO_EXECUTABLE|g" "$SERVICE_FILE_TEMPLATE" > "$SERVICE_FILE"

# Copy the systemd service file to the systemd directory
echo "Copying systemd service file..."
sudo cp "$SERVICE_FILE" "$SYSTEMD_DIR"

# Unmask the service if it is masked
if systemctl list-unit-files | grep -q "$SERVICE_NAME.*masked"; then
  echo "Unmasking $SERVICE_NAME..."
  sudo systemctl unmask "$SERVICE_NAME"
fi

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
