#!/bin/bash

# Load environment variables from a .env file
source ./.env

REMOTE_USER="$REMOTE_USER"
REMOTE_HOST="$SERVER_IP"
REMOTE_DIR="/home/$REMOTE_USER/expo-build-service"
REMOTE_PASSWORD="$REMOTE_PASSWORD"

# Update package lists
echo "Updating package lists..."
sudo swupd update
sudo swupd diagnose
sudo swupd repair
flatpak update

# Install required packages
echo "Installing required packages..."
sudo swupd bundle-add nodejs-basic go-basic gh

# Install eas-cli globally using npm
if ! command -v eas &> /dev/null; then
  echo "Installing eas-cli..."
  npm install -g eas-cli
fi

# Install Java (Amazon Corretto 17)
echo "Installing Java (Amazon Corretto 17)..."
# Ensure the script is executable
chmod +x install_corretto_server.sh
# Run the Java installation script
./install_corretto_server.sh

# Install Android SDK
echo "Installing Android SDK..."
chmod +x setup_android_sdk_server.sh
./setup_android_sdk_server.sh

source .bashrc

# Create a tarball of the contents of the current directory in a temporary directory
TEMP_DIR=$(mktemp -d)
TARBALL="$TEMP_DIR/expo-build-service.tar.gz"

CURRENT_DIR="$(pwd)"

# Tar the contents of the current directory without including the directory itself
tar --exclude="$TARBALL" -czf "$TARBALL" -C "$CURRENT_DIR" .

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
  # Check for a key file to ensure extraction was successful
  if [ ! -f "install_server.sh" ]; then
    echo "Error: install_server.sh does not exist after extraction"
    exit 1
  fi
  chmod +x install_server.sh
  echo "$REMOTE_PASSWORD" | sudo -S ./install_server.sh
EOF

# Clean up the local tarball
rm -rf "$TEMP_DIR"
