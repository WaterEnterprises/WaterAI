package process

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"water-ai/core"
)

// ManagerConfig holds the configuration for the process manager
type ManagerConfig struct {
	GatewayPort    string // Port for the gateway server (e.g., "8080")
	FrontendPort   string // Port for the frontend server (e.g., "8000")
	GatewayPath    string // Path to the gateway binary or command
	HealthInterval time.Duration
	MaxRestartAttempts int
	RestartDelay    time.Duration
	MaxRestartDelay time.Duration
}

// Manager handles the lifecycle of the gateway process
type Manager struct {
	config       ManagerConfig
	cmd          *exec.Cmd
	cmdMu        sync.RWMutex
	isRunning    bool
	lastCheck    time.Time
	statusMu     sync.RWMutex
	httpClient   *http.Client
	logger       *slog.Logger
	stopChan     chan struct{}
	doneChan     chan struct{}
	restartCount int
	restartMu    sync.Mutex
}

// NewManager creates a new process manager
func NewManager(cfg ManagerConfig) *Manager {
	if cfg.HealthInterval == 0 {
		cfg.HealthInterval = 5 * time.Second
	}
	if cfg.MaxRestartAttempts == 0 {
		cfg.MaxRestartAttempts = 10
	}
	if cfg.RestartDelay == 0 {
		cfg.RestartDelay = 1 * time.Second
	}
	if cfg.MaxRestartDelay == 0 {
		cfg.MaxRestartDelay = 30 * time.Second
	}

	return &Manager{
		config: cfg,
		logger: core.Logger.With("component", "process_manager"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

// Start begins the gateway process and starts monitoring
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("starting process manager",
		"gateway_port", m.config.GatewayPort,
		"frontend_port", m.config.FrontendPort)

	// Start the gateway process
	if err := m.startGateway(ctx); err != nil {
		return fmt.Errorf("failed to start gateway: %w", err)
	}

	// Start the health check loop
	go m.healthCheckLoop(ctx)

	m.logger.Info("process manager started successfully")
	return nil
}

// startGateway launches the gateway as a child process
func (m *Manager) startGateway(ctx context.Context) error {
	m.cmdMu.Lock()
	defer m.cmdMu.Unlock()

	// Create the command - run the same binary with gateway mode
	m.cmd = exec.CommandContext(ctx, os.Args[0], "gateway",
		"--port", m.config.GatewayPort,
	)

	// Set process group so we can terminate all child processes
	m.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:   0,
	}

	// Capture output
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	m.logger.Info("spawning gateway process", "command", m.cmd.String())

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gateway process: %w", err)
	}

	m.statusMu.Lock()
	m.isRunning = true
	m.statusMu.Unlock()

	m.logger.Info("gateway process started", "pid", m.cmd.Process.Pid)
	return nil
}

// healthCheckLoop continuously monitors the gateway process
func (m *Manager) healthCheckLoop(ctx context.Context) {
	m.logger.Info("starting health check loop", "interval", m.config.HealthInterval.String())

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("context cancelled, stopping health check loop")
			close(m.doneChan)
			return
		case <-m.stopChan:
			m.logger.Info("stop signal received, stopping health check loop")
			close(m.doneChan)
			return
		default:
			if !m.checkHealth() {
				m.logger.Warn("gateway health check failed, attempting restart")
				if err := m.restartGateway(ctx); err != nil {
					m.logger.Error("failed to restart gateway", "error", err)
				}
			}
			time.Sleep(m.config.HealthInterval)
		}
	}
}

// checkHealth verifies the gateway is responsive
func (m *Manager) checkHealth() bool {
	m.statusMu.RLock()
	isRunning := m.isRunning
	m.statusMu.RUnlock()

	if !isRunning {
		return false
	}

	// Check if process is still alive
	m.cmdMu.RLock()
	cmd := m.cmd
	m.cmdMu.RUnlock()

	if cmd.Process == nil {
		return false
	}

	// Try to ping the health endpoint
	url := fmt.Sprintf("http://localhost:%s/health", m.config.GatewayPort)
	resp, err := m.httpClient.Get(url)
	if err != nil {
		m.logger.Debug("health check failed", "url", url, "error", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Warn("health check returned non-OK status", "status", resp.StatusCode)
		return false
	}

	m.lastCheck = time.Now()
	return true
}

// restartGateway attempts to restart the gateway process with exponential backoff
func (m *Manager) restartGateway(ctx context.Context) error {
	m.restartMu.Lock()
	m.restartCount++
	m.restartMu.Unlock()

	// Calculate exponential backoff delay
	m.restartMu.Lock()
	delay := time.Duration(m.restartCount) * m.config.RestartDelay
	if delay > m.config.MaxRestartDelay {
		delay = m.config.MaxRestartDelay
	}
	// Reset restart count after max delay
	if m.restartCount >= m.config.MaxRestartAttempts {
		m.restartCount = 0
	}
	m.restartMu.Unlock()

	m.logger.Info("attempting to restart gateway",
		"attempt", m.restartCount,
		"delay", delay.String())

	// Stop the current process
	if err := m.stopGateway(); err != nil {
		m.logger.Warn("error stopping gateway during restart", "error", err)
	}

	// Wait before restarting (with exponential backoff)
	time.Sleep(delay)

	// Start a new process
	return m.startGateway(ctx)
}

// stopGateway gracefully stops the gateway process
func (m *Manager) stopGateway() error {
	m.cmdMu.Lock()
	cmd := m.cmd
	m.cmdMu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	m.logger.Info("stopping gateway process", "pid", cmd.Process.Pid)

	m.statusMu.Lock()
	m.isRunning = false
	m.statusMu.Unlock()

	// Send SIGTERM
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		m.logger.Warn("failed to send SIGTERM, sending SIGKILL", "error", err)
		// Force kill if SIGTERM fails
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			m.logger.Debug("process exited with error", "error", err)
		}
	case <-time.After(10 * time.Second):
		m.logger.Warn("process did not exit in time, forcing kill")
		cmd.Process.Kill()
	}

	m.logger.Info("gateway process stopped")
	return nil
}

// Stop gracefully shuts down the process manager and all child processes
func (m *Manager) Stop() error {
	close(m.stopChan)

	// Wait for health check loop to finish
	select {
	case <-m.doneChan:
	case <-time.After(5 * time.Second):
		m.logger.Warn("health check loop did not stop in time")
	}

	// Stop the gateway
	return m.stopGateway()
}

// IsRunning returns whether the gateway is currently running
func (m *Manager) IsRunning() bool {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()
	return m.isRunning
}

// GetGatewayPort returns the configured gateway port
func (m *Manager) GetGatewayPort() string {
	return m.config.GatewayPort
}

// GetFrontendPort returns the configured frontend port
func (m *Manager) GetFrontendPort() string {
	return m.config.FrontendPort
}
