#!/bin/bash

# Ensure the script is executed from the correct directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || { echo "Failed to change directory to $SCRIPT_DIR"; exit 1; }

# Path to the update log
LOG_FILE="$SCRIPT_DIR/logs/server.log"

# Log the start of the update process
echo "$(date '+%Y-%m-%d %H:%M:%S') - Starting server update..." | tee -a "$LOG_FILE"

# Pull the latest code
echo "$(date '+%Y-%m-%d %H:%M:%S') - Pulling latest code..." | tee -a "$LOG_FILE"
git fetch --all
git reset --hard origin/main  # Replace 'main' with your default branch if different

# Build the new Go executable
echo "$(date '+%Y-%m-%d %H:%M:%S') - Building Go executable..." | tee -a "$LOG_FILE"
go build -o buildHandler .

# Restart the server using systemd
echo "$(date '+%Y-%m-%d %H:%M:%S') - Restarting go-server.service..." | tee -a "$LOG_FILE"
sudo systemctl restart go-server.service

# Confirm the update completion
echo "$(date '+%Y-%m-%d %H:%M:%S') - Server updated and restarted." | tee -a "$LOG_FILE"
exit 0
