# Expo Build Service

This project provides a Go-based service for building Expo applications. It supports building for both Android and iOS platforms and includes features for cloning repositories, running npm install, and building the application using the EAS CLI. The service also includes an update mechanism and health check endpoint.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Endpoints](#endpoints)
- [Running on a Remote Server](#running-on-a-remote-server)
- [License](#license)

## Prerequisites

- Go (1.16+)
- Git
- npm
- EAS CLI
- Systemd (for Linux systems)
- SSH access to the server (for remote installation)

## Installation

1. **Clone the repository:**

    ```sh
    git clone https://github.com/yourusername/expo-build-service.git
    cd expo-build-service
    ```

2. **Create a `.env` file:**

    Create a `.env` file in the `expo-build-service` directory with the following content:

    ```env
    AUTH_TOKEN=your-secret-token
    SERVER_IP=your-server-ip
    ```

3. **Run the installation script:**

    ```sh
    ./install_server.sh
    ```

    During the installation, you will be prompted to enter the user and group to run the service.

## Configuration

The service uses environment variables for configuration. The following variables are required:

- `AUTH_TOKEN`: The token used for authenticating requests.
- `SERVER_IP`: The IP address of the server.

These variables should be set in the `.env` file located in the `expo-build-service` directory.

## Usage

### Building and Downloading APK

To build and download an APK, you can use the `script.sh` file:

```sh
./script.sh
```

This script will trigger the build process and download the APK file.

### Triggering Server Update

To trigger a server update, you can uncomment the `trigger_update` function call in the `script.sh` file and run the script:

```sh
# Uncomment the following line in script.sh
# trigger_update
./script.sh
```

## Endpoints

### `/build`

- **Method:** `POST`
- **Description:** Triggers the build process for the specified repository and platform.
- **Request Body:**
    ```json
    {
        "repo_url": "https://github.com/yourusername/your-repo.git",
        "platform": "android",
        "package_path": "path/to/package"
    }
    ```
If the `package_path` is not provided, the default path will be used.
- **Headers:**
    - `Authorization: Bearer your-secret-token`

### `/update`

- **Method:** `GET`
- **Description:** Triggers the server update process.
- **Headers:**
    - `Authorization: Bearer your-secret-token`

### `/health`

- **Method:** `GET`
- **Description:** Checks the health of the server.

## Running on a Remote Server

To run the installation script on a remote server using SSH, you can use the following script:

```sh
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
```

## License

This project is licensed under the GNU GENERAL PUBLIC LICENSE. See the [LICENSE](LICENSE) file for details.
