package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"water-ai/core"
	"water-ai/resources"
	"water-ai/server"
	"water-ai/ui"
	"water-ai/ui/theme"

	"fyne.io/fyne/v2/app"
)

// Build-time variables injected via ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

const (
	serverPort = "7777"
	serverURL  = "http://localhost:" + serverPort
	healthURL  = serverURL + "/health"
)

func main() {
	// ---------------------------------------------------------
	// MODE 1: Background Service (headless daemon)
	// ---------------------------------------------------------
	// When invoked with "server" argument, run only the gateway.
	if len(os.Args) > 1 && os.Args[1] == "server" {
		runBackgroundService()
		return
	}

	// ---------------------------------------------------------
	// MODE 2: Version flag
	// ---------------------------------------------------------
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Water AI %s (commit: %s, built: %s, go: %s)\n",
			Version, GitCommit, BuildDate, GoVersion)
		return
	}

	// ---------------------------------------------------------
	// MODE 3: Unified GUI + Gateway (default)
	// ---------------------------------------------------------
	// Start the gateway in a background goroutine, then launch
	// the Fyne GUI on the main thread. When the GUI window is
	// closed the gateway is gracefully shut down.
	runUnified()
}

// runUnified starts the gateway service in a goroutine and the Fyne GUI
// on the main thread. The gateway is shut down when the GUI exits.
func runUnified() {
	logger := core.Logger

	// --- Start the gateway server in the background ---
	srv := server.CreateServer(server.Config{
		Port: serverPort,
	})

	// Add health endpoint for connectivity checks
	srv.Router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	httpServer := &http.Server{
		Addr:    ":" + serverPort,
		Handler: srv.Router,
	}

	go func() {
		logger.Info("Water AI gateway starting", "port", serverPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Gateway failed to start", "error", err)
		}
	}()

	// Wait briefly for the server to be ready
	waitForServer(healthURL, 5*time.Second)

	// --- Launch the Fyne GUI on the main thread ---
	a := app.NewWithID("com.waterai.gui")
	a.Settings().SetTheme(theme.NewWaterAITheme())
	a.SetIcon(resources.GetLogoOnly())

	mainWindow := ui.NewMainWindow(a)

	// Show the window and run the event loop (blocks until quit)
	mainWindow.ShowAndRun()

	// --- GUI has exited — shut down the gateway gracefully ---
	logger.Info("GUI closed, shutting down gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Gateway shutdown error", "error", err)
	}
	logger.Info("Water AI shut down cleanly")
}

// runBackgroundService runs the gateway as a standalone headless service.
func runBackgroundService() {
	logger := core.Logger
	logger.Info("Water AI Background Service Started", "port", serverPort)

	srv := server.CreateServer(server.Config{
		Port: serverPort,
	})

	srv.Router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	if err := srv.Router.Run(":" + serverPort); err != nil {
		logger.Error("Service failed to start", "error", err)
		os.Exit(1)
	}
}

// waitForServer polls the health endpoint until it responds or the timeout elapses.
func waitForServer(url string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	client := http.Client{Timeout: 500 * time.Millisecond}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
