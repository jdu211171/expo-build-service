#!/bin/bash

# Ensure the script is executed from the correct directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || { echo "Failed to change directory to $SCRIPT_DIR"; exit 1; }

# Variables
GO_EXECUTABLE="buildHandler"

# Run the Go executable
echo "Running Go executable..."
./"$GO_EXECUTABLE"
