package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"water-ai/core"
	"water-ai/server"
)

const (
	serverPort = "7777"
	serverURL  = "http://localhost:" + serverPort
	healthURL  = serverURL + "/health"
)




func main() {
	// ---------------------------------------------------------
	// MODE 1: Background Service (The "Daemon")
	// ---------------------------------------------------------
	// This block only runs when the binary is spawned with the "server" argument.
	if len(os.Args) > 1 && os.Args[1] == "server" {
		runBackgroundService()
		return
	}

	// ---------------------------------------------------------
	// MODE 2: Launcher (The "Client" / User Click)
	// ---------------------------------------------------------
	// This runs when you click the binary or run it in terminal.
	runLauncherLogic()
}

// runLauncherLogic checks for the service and starts it if missing
func runLauncherLogic() {
	// 1. Check if the service is already alive
	if isServerHealthy() {
		fmt.Println("Water AI Service is already running.")
		fmt.Println("Opening browser...")
		_ = openBrowser(serverURL)
		return
	}

	// 2. Service is dead. We need to spawn it.
	fmt.Println("Water AI Service is stopped. Starting background process...")
	if err := spawnBackgroundProcess(); err != nil {
		fmt.Printf("Fatal Error: Could not spawn service: %v\n", err)
		os.Exit(1)
	}

	// 3. Wait for the background service to wake up
	fmt.Print("Waiting for initialization...")
	for i := 0; i < 20; i++ { // Wait up to 10 seconds (20 * 500ms)
		if isServerHealthy() {
			fmt.Println("\nSuccess! Water AI is ready.")
			fmt.Println("Opening browser...")
			_ = openBrowser(serverURL)
			return
		}
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
	}

	fmt.Println("\nError: Timed out waiting for Water AI to start.")
	fmt.Println("Please check logs or try running 'water-ai server' manually to debug.")
	os.Exit(1)
}

// runBackgroundService is the actual blocking web server
func runBackgroundService() {
	// Setup Logging
	logger := core.Logger
	logger.Info("Water AI Background Service Started", "port", serverPort)

	// Initialize the Server (API + Static Files + WebSocket)
	srv := server.CreateServer(server.Config{
		Port: serverPort,
	})

	// Ensure we have a specific health check endpoint for the launcher
	srv.Router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Run blocking
	if err := srv.Router.Run(":" + serverPort); err != nil {
		logger.Error("Service failed to start", "error", err)
		os.Exit(1)
	}
}

// --- Helpers ---

// spawnBackgroundProcess starts a new instance of this binary with "server" arg
func spawnBackgroundProcess() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable path: %w", err)
	}

	// Spawn ourselves with the "server" argument
	cmd := exec.Command(exe, "server")

	// Detach from current console (Uses logic from process_windows.go / process_unix.go)
	configureDetachedProcess(cmd)

	// Ensure no I/O binding prevents detachment
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Start()
}

// isServerHealthy checks if port 7777 is responding
func isServerHealthy() bool {
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// openBrowser opens the system default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}