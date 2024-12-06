package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort         string
	LogDirectory       string
	LogFile            string
	BuildTimeout       time.Duration
	TempDirPrefix      string
	UpdateScriptPath   string
	AllowedPlatforms   []string
	DefaultCloneBranch string
}

// Load configuration from environment variables
func loadConfig() Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default configuration")
	}

	return Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		LogDirectory:       getEnv("LOG_DIRECTORY", "/home/server/expo-build-service/logs"),
		LogFile:            getEnv("LOG_FILE", "server.log"),
		BuildTimeout:       parseDuration(getEnv("BUILD_TIMEOUT", "60m")),
		TempDirPrefix:      getEnv("TEMP_DIR_PREFIX", "build-"),
		UpdateScriptPath:   getEnv("UPDATE_SCRIPT_PATH", "/home/server/expo-build-service/update_server.sh"),
		AllowedPlatforms:   strings.Split(getEnv("ALLOWED_PLATFORMS", "android,ios"), ","),
		DefaultCloneBranch: getEnv("DEFAULT_CLONE_BRANCH", "main"),
	}
}

// Helper function to get environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to parse duration safely
func parseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Printf("Invalid duration %s, using default 60 minutes", durationStr)
		return 60 * time.Minute
	}
	return duration
}

// BuildRequest defines the expected JSON payload for build requests
type BuildRequest struct {
	RepoURL      string `json:"repo_url"`
	Platform     string `json:"platform"`
	PackagePath  string `json:"package_path"`
	UpdateServer bool   `json:"update_server"`
}

// Modify handlers and main function to use config
func buildHandler(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), config.BuildTimeout)
		defer cancel()

		var req BuildRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("Invalid request payload:", err)
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Validate input
		if req.RepoURL == "" || req.Platform == "" || req.PackagePath == "" {
			log.Println("Missing required parameters")
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}

		// Rest of the existing buildHandler logic,
		// passing config where needed
		// ... (keep the existing implementation, just modify to use config)
		// Proceed with the build logic
		buildID := generateTimestampID()

		// Create a temporary directory for this build
		tempDir, err := os.MkdirTemp("", "build-"+buildID)
		if err != nil {
			log.Println("Failed to create temporary directory:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer func(path string) {
			err := os.RemoveAll(path)
			if err != nil {
				log.Printf("Failed to clean up temporary directory %s: %v", path, err)
			}
		}(tempDir) // Clean up after build

		clonePath := filepath.Join(tempDir, "repo")

		// Clone the repository
		if err := cloneOrUpdateRepo(ctx, req.RepoURL, clonePath); err != nil {
			log.Println("Failed to clone the repository:", err)
			http.Error(w, "Failed to clone the repository", http.StatusInternalServerError)
			return
		}

		// Run npm install in the package directory
		packagePath := filepath.Join(clonePath, req.PackagePath)
		if err := runNpmInstall(ctx, packagePath); err != nil {
			log.Println("Failed to install npm dependencies:", err)
			http.Error(w, "Failed to install npm dependencies", http.StatusInternalServerError)
			return
		}

		// Define the output file based on the platform and build ID
		var outputFile, contentType, outputFilename string
		switch req.Platform {
		case "android":
			outputFilename = fmt.Sprintf("app-%s.apk", buildID)
			outputFile = outputFilename
			contentType = "application/vnd.android.package-archive"
		case "ios":
			outputFilename = fmt.Sprintf("app-%s.ipa", buildID)
			outputFile = outputFilename
			contentType = "application/octet-stream"
		default:
			log.Println("Unsupported platform:", req.Platform)
			http.Error(w, "Unsupported platform", http.StatusBadRequest)
			return
		}

		// Tail the log file
		done := make(chan struct{})
		go tailLogFile(w, "/home/server/expo-build-service/logs/server.log", done)

		// Build the app
		if err := buildApp(ctx, packagePath, req.Platform, outputFile); err != nil {
			log.Println("Failed to build the app:", err)
			http.Error(w, "Failed to build the app", http.StatusInternalServerError)
			close(done)
			return
		}

		// Serve the built app
		builtFilePath := filepath.Join(packagePath, outputFile)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", outputFilename))
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize(builtFilePath)))

		// Stream the file to the client
		file, err := os.Open(builtFilePath)
		if err != nil {
			log.Println("Failed to open built file:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			close(done)
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)

		if _, err := io.Copy(w, file); err != nil {
			log.Println("Failed to send file to client:", err)
		}

		// Stop tailing the log file
		close(done)
	}
}

func updateHandler(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authenticate the request
		token := r.Header.Get("Authorization")
		expectedToken := os.Getenv("UPDATE_AUTH_TOKEN")
		if token != "Bearer "+expectedToken {
			log.Println("Unauthorized access attempt")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Rest of the existing updateHandler logic
		// Use config.UpdateScriptPath instead of hardcoded path
		go func() {
			cmd := exec.Command(config.UpdateScriptPath)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Update failed: %v\nOutput: %s", err, string(output))
			} else {
				log.Println("Update completed successfully.")
			}
		}()

		// ... (remaining logic)
	}
}

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize logging with config
	initLogging(config)

	srv := &http.Server{
		Addr: "0.0.0.0:" + config.ServerPort,
	}

	// Register handlers with config
	http.HandleFunc("/build", authenticate(buildHandler(config)))
	http.HandleFunc("/update", updateHandler(config))
	http.HandleFunc("/health", healthHandler)

	// Start the server
	go func() {
		log.Printf("Server started at :%s", config.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown logic remains the same
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// Modify initLogging to use config
func initLogging(config Config) {
	logDir := config.LogDirectory
	logFile := filepath.Join(logDir, config.LogFile)

	// Create log directory if it doesn't exist
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	// Open log file in append mode, create if not exists
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file %s: %v", logFile, err)
	}

	// Set log output to the file
	log.SetOutput(file)

	// Set log flags to include date and time
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Health check handler
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintln(w, "Server is up and running.")
	if err != nil {
		return
	}
}

// Authentication middleware
func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := os.Getenv("AUTH_TOKEN")
		// Log the tokens for debugging
		log.Printf("Received token: %s", token)
		log.Printf("Expected token: Bearer %s", expectedToken)
		if token != "Bearer "+expectedToken {
			log.Println("Unauthorized access attempt")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func buildApp(ctx context.Context, packagePath, platform, outputFile string) error {
	// Validate the platform
	validPlatforms := map[string]bool{"android": true, "ios": true}
	if !validPlatforms[platform] {
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	// Build the app using EAS CLI
	buildCmd := exec.CommandContext(ctx, "eas", "build", "--platform", platform, "--local", "--output", outputFile)
	buildCmd.Dir = packagePath
	buildCmd.Env = os.Environ() // Inherit the environment

	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error building app: %v, output: %s", err, string(output))
	}

	// Check if the built file exists
	builtFilePath := filepath.Join(packagePath, outputFile)
	if _, err := os.Stat(builtFilePath); os.IsNotExist(err) {
		return fmt.Errorf("built app file not found at %s", builtFilePath)
	}

	return nil
}

func fileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		log.Println("Failed to get file size:", err)
		return 0
	}
	return info.Size()
}

// Tail the log file and send updates to the client
func tailLogFile(w http.ResponseWriter, logFilePath string, done chan struct{}) {
	cmd := exec.Command("tail", "-f", logFilePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Failed to get stdout pipe:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Println("Failed to start tail command:", err)
		return
	}

	reader := bufio.NewReader(stdout)
	for {
		select {
		case <-done:
			err := cmd.Process.Kill()
			if err != nil {
				return
			}
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if _, err := w.Write([]byte(line)); err != nil {
				log.Println("Failed to send log message:", err)
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// Run npm install in the specified package directory
func runNpmInstall(ctx context.Context, packagePath string) error {
	installCmd := exec.CommandContext(ctx, "npm", "install")
	installCmd.Dir = packagePath
	installCmd.Env = os.Environ() // Inherit the environment

	if output, err := installCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error running npm install: %v, output: %s", err, string(output))
	}

	return nil
}

// Clone or update the repository
func cloneOrUpdateRepo(ctx context.Context, repoURL, clonePath string) error {
	if strings.ContainsAny(repoURL, ";&") {
		return fmt.Errorf("invalid repoURL parameter")
	}

	// Create the parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(clonePath), 0755); err != nil {
		return fmt.Errorf("error creating parent directory: %v", err)
	}

	// Perform a shallow clone of the main branch
	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--single-branch", "--branch", "main", repoURL, clonePath)

	// Set the GIT_TERMINAL_PROMPT environment variable to prevent interactive prompts
	cloneCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	// Use a buffer to capture output
	var output bytes.Buffer
	cloneCmd.Stdout = &output
	cloneCmd.Stderr = &output

	// Run the command
	err := cloneCmd.Run()
	if err != nil {
		return fmt.Errorf("error cloning repository: %v, output: %s", err, output.String())
	}

	return nil
}

// Generate a timestamp-based ID for builds
func generateTimestampID() string {
	timestamp := time.Now().Format("20060102-1504") // YearMonthDay-HourMinute
	return timestamp
}
