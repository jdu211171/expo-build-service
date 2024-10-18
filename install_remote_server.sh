#!/bin/bash

# Load environment variables from a .env file
source ./.env

REMOTE_USER="$REMOTE_USER"
REMOTE_HOST="$SERVER_IP"
REMOTE_DIR="/home/$REMOTE_USER/expo-build-service"
REMOTE_PASSWORD="$REMOTE_PASSWORD"

# Create a tarball of the current directory in a temporary directory
TEMP_DIR=$(mktemp -d)
TARBALL="$TEMP_DIR/expo-build-service.tar.gz"

CURRENT_DIR="$(pwd)"
CURRENT_DIR_NAME="$(basename "$CURRENT_DIR")"  # Should be expo-build-service
PARENT_DIR="$(dirname "$CURRENT_DIR")"

cd "$PARENT_DIR"
tar --exclude="$TARBALL" -czf "$TARBALL" "$CURRENT_DIR_NAME"
cd "$CURRENT_DIR"  # Go back to the original directory

# Ensure the remote directory exists
sshpass -p "$REMOTE_PASSWORD" ssh "$REMOTE_USER@$REMOTE_HOST" "mkdir -p $REMOTE_DIR"

# Copy the tarball to the remote server
sshpass -p "$REMOTE_PASSWORD" scp "$TARBALL" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR"

# Run the install script on the remote server
sshpass -p "$REMOTE_PASSWORD" ssh "$REMOTE_USER@$REMOTE_HOST" << EOF
  set -e  # Exit immediately if a command exits with a non-zero status
  cd $REMOTE_DIR
  tar -xzf $(basename $TARBALL)
  rm $(basename $TARBALL)
  if [ ! -d "expo-build-service" ]; then
    echo "Error: Directory expo-build-service does not exist after extraction"
    exit 1
  fi
  cd expo-build-service
  chmod +x install_server.sh
  echo "$REMOTE_PASSWORD" | sudo -S ./install_server.sh
EOF

# Clean up the local tarball
rm -rf "$TEMP_DIR"
