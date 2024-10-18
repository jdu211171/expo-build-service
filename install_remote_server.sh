#!/bin/bash

# Load environment variables from a .env file
source /path/to/.env

REMOTE_USER="<username>"
REMOTE_HOST="$SERVER_IP"
REMOTE_DIR="/home/$REMOTE_USER/Go/expo-build-service"

# Copy the project files to the remote server
scp -r . "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR"

# Run the install script on the remote server
ssh "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_DIR && ./install_server.sh"
