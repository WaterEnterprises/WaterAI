package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"water-ai/server"
)

// TestGatewayStartsAndHealthy is a smoke test that verifies the gateway
// service starts up and responds to health checks on a random port.
func TestGatewayStartsAndHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use a test port to avoid conflicts
	testPort := "17777"
	srv := server.CreateServer(server.Config{
		Port: testPort,
	})

	srv.Router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	httpServer := &http.Server{
		Addr:    ":" + testPort,
		Handler: srv.Router,
	}

	// Start the server in a goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()

	// Wait for the server to be ready
	healthURL := "http://localhost:" + testPort + "/health"
	client := http.Client{Timeout: 500 * time.Millisecond}

	var healthy bool
	for i := 0; i < 20; i++ {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				healthy = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !healthy {
		t.Fatal("gateway did not become healthy within timeout")
	}

	// Verify the API routes are registered
	apiURL := "http://localhost:" + testPort + "/api/settings"
	resp, err := client.Get(apiURL)
	if err != nil {
		t.Fatalf("failed to reach /api/settings: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from /api/settings, got %d", resp.StatusCode)
	}

	// Shut down
	httpServer.Close()
}

// TestWaitForServer verifies the waitForServer helper.
func TestWaitForServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testPort := "17778"
	srv := server.CreateServer(server.Config{Port: testPort})
	srv.Router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	httpServer := &http.Server{
		Addr:    ":" + testPort,
		Handler: srv.Router,
	}

	go func() {
		httpServer.ListenAndServe()
	}()
	defer httpServer.Close()

	// waitForServer should return without error
	healthURL := "http://localhost:" + testPort + "/health"
	waitForServer(healthURL, 5*time.Second)

	// Verify it's actually reachable
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		t.Fatalf("server not reachable after waitForServer: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
