package process

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"water-ai/core"
	"water-ai/server"
)

// Gateway wraps the server for use as a child process
type Gateway struct {
	config    GatewayConfig
	logger    *slog.Logger
	server    *server.Server
	router    *gin.Engine
	httpServer *http.Server
	mu       sync.RWMutex

	// Graceful shutdown
	stopChan chan struct{}
	doneChan chan struct{}
}

// NewGateway creates a new gateway instance
func NewGateway(cfg GatewayConfig) *Gateway {
	return &Gateway{
		config:   cfg,
		logger:   core.Logger.With("component", "gateway"),
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

// Start runs the gateway server
func (g *Gateway) Start(ctx context.Context) error {
	g.logger.Info("starting gateway server",
		"port", g.config.Port,
		"workspace", os.Getenv("WORKSPACE_ROOT"),
	)

	// Create server config
	serverConfig := server.Config{
		WorkspaceRoot: os.Getenv("WORKSPACE_ROOT"),
		Port:          g.config.Port,
	}

	// Create the server
	g.server = server.CreateServer(serverConfig)
	g.router = g.server.Router

	// Add health check endpoint
	g.router.GET("/health", g.healthHandler)
	g.router.GET("/ready", g.readyHandler)

	// Create HTTP server
	g.httpServer = &http.Server{
		Addr:    ":" + g.config.Port,
		Handler: g.router,
	}

	// Start server in goroutine
	go func() {
		g.logger.Info("gateway server listening", "address", g.httpServer.Addr)
		if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.logger.Error("gateway server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	g.waitForShutdown(ctx)

	g.logger.Info("gateway server stopped")
	return nil
}

// healthHandler returns the health status of the gateway
func (g *Gateway) healthHandler(c *gin.Context) {
	g.mu.RLock()
	running := g.server != nil
	g.mu.RUnlock()

	if !running {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"message": "server not initialized",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"port":      g.config.Port,
	})
}

// readyHandler returns the readiness status of the gateway
func (g *Gateway) readyHandler(c *gin.Context) {
	g.mu.RLock()
	ready := g.server != nil
	g.mu.RUnlock()

	if !ready {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// waitForShutdown blocks until a shutdown signal is received or context is cancelled
func (g *Gateway) waitForShutdown(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ctx.Done():
			g.logger.Info("context cancelled, shutting down gateway")
			g.shutdown()
			return
		case sig := <-sigChan:
			g.logger.Info("received shutdown signal", "signal", sig.String())
			g.shutdown()
			return
		case <-g.stopChan:
			g.logger.Info("stop signal received, shutting down gateway")
			g.shutdown()
			return
		}
	}
}

// shutdown gracefully stops the gateway server
func (g *Gateway) shutdown() {
	g.logger.Info("initiating graceful shutdown")

	g.mu.Lock()
	if g.httpServer != nil {
		// Create a timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := g.httpServer.Shutdown(ctx); err != nil {
			g.logger.Error("error shutting down http server", "error", err)
		}
	}
	g.mu.Unlock()

	close(g.doneChan)
}

// Stop sends a stop signal to the gateway
func (g *Gateway) Stop() {
	close(g.stopChan)

	// Wait for shutdown to complete
	select {
	case <-g.doneChan:
	case <-time.After(10 * time.Second):
		g.logger.Warn("shutdown did not complete in time")
	}
}

// GetPort returns the configured port
func (g *Gateway) GetPort() string {
	return g.config.Port
}

// GatewayConfig holds the gateway configuration
type GatewayConfig struct {
	Port          string
	WorkspaceRoot string
}

// GetGatewayConfig returns the configuration for the gateway process
func GetGatewayConfig() GatewayConfig {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	workspace := os.Getenv("WORKSPACE_ROOT")
	if workspace == "" {
		workspace = "./workspace_data"
	}

	return GatewayConfig{
		Port:          port,
		WorkspaceRoot: workspace,
	}
}

// RunGateway is the entry point when running as a gateway process
func RunGateway() {
	cfg := GetGatewayConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gateway := NewGateway(GatewayConfig{
		Port:          cfg.Port,
		WorkspaceRoot: cfg.WorkspaceRoot,
	})

	gateway.logger.Info("starting as gateway process", "port", cfg.Port)

	if err := gateway.Start(ctx); err != nil {
		gateway.logger.Error("gateway failed", "error", err)
		os.Exit(1)
	}
}

// CreateGatewayManagerConfig creates a Manager config for running the gateway
func CreateGatewayManagerConfig(gatewayPort, frontendPort string) ManagerConfig {
	return ManagerConfig{
		GatewayPort:       gatewayPort,
		FrontendPort:      frontendPort,
		HealthInterval:    5 * time.Second,
		MaxRestartAttempts: 5,
		RestartDelay:       2 * time.Second,
	}
}

// IsGatewayMode checks if the process is running in gateway mode
func IsGatewayMode() bool {
	for i, arg := range os.Args {
		if arg == "gateway" && i+1 < len(os.Args) {
			return true
		}
	}
	return false
}
