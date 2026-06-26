package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// --- Docker Sandbox ---

type DockerSandbox struct {
	Base
	client *client.Client
}

func init() {
	Register(ModeDocker, func(sid string, s *Settings) Sandbox {
		cli, _ := client.NewClientWithOpts(client.FromEnv)
		return &DockerSandbox{Base: Base{SessionID: sid, Settings: s}, client: cli}
	})
}

func (s *DockerSandbox) Create(ctx context.Context) error {
	// Simplified volume and config logic
	config := &container.Config{
		Image: s.Settings.SandboxConfig.Image,
		Tty:   true,
	}
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   s.Settings.SandboxConfig.MemoryLimit,
			NanoCPUs: int64(s.Settings.SandboxConfig.CPULimit * 1e9),
		},
	}

	resp, err := s.client.ContainerCreate(ctx, config, hostConfig, nil, nil, s.SessionID)
	if err != nil {
		return err
	}

	s.SandboxID = resp.ID
	// Fixed: types.ContainerStartOptions is now container.StartOptions
	if err := s.client.ContainerStart(ctx, s.SandboxID, container.StartOptions{}); err != nil {
		return err
	}

	s.HostURL = fmt.Sprintf("http://%s:%d", s.SessionID, s.Settings.SandboxConfig.ServicePort)
	return nil
}

func (s *DockerSandbox) Connect(ctx context.Context) error {
	s.HostURL = fmt.Sprintf("http://%s:%d", s.SessionID, s.Settings.SandboxConfig.ServicePort)
	return nil
}

func (s *DockerSandbox) Start(ctx context.Context) error   { return nil }
func (s *DockerSandbox) Stop(ctx context.Context) error    { return s.client.ContainerStop(ctx, s.SandboxID, container.StopOptions{}) }
func (s *DockerSandbox) Cleanup(ctx context.Context) error { return s.client.ContainerRemove(ctx, s.SandboxID, container.RemoveOptions{Force: true}) }
func (s *DockerSandbox) ExposePort(port int) string        { return fmt.Sprintf("http://%s-%d.%s", s.SessionID, port, os.Getenv("BASE_URL")) }

// --- E2B Sandbox (Mocked/Simplified for Go) ---

type E2BSandbox struct {
	Base
}

func init() {
	Register(ModeE2B, func(sid string, s *Settings) Sandbox {
		return &E2BSandbox{Base: Base{SessionID: sid, Settings: s}}
	})
}

func (s *E2BSandbox) Create(ctx context.Context) error {
	// In a real scenario, you'd use an E2B Go SDK or HTTP client here
	s.SandboxID = "e2b-proto-" + s.SessionID
	s.HostURL = s.ExposePort(s.Settings.SandboxConfig.ServicePort)
	return nil
}

func (s *E2BSandbox) Connect(ctx context.Context) error { return nil }
func (s *E2BSandbox) Start(ctx context.Context) error   { return nil }
func (s *E2BSandbox) Stop(ctx context.Context) error    { return nil }
func (s *E2BSandbox) Cleanup(ctx context.Context) error { return nil }
func (s *E2BSandbox) ExposePort(port int) string {
	return fmt.Sprintf("https://%d-%s.e2b.dev", port, s.SandboxID)
}

// --- Local Sandbox ---

type LocalSandbox struct {
	Base
	cmd *exec.Cmd
}

func init() {
	Register(ModeLocal, func(sid string, s *Settings) Sandbox {
		return &LocalSandbox{Base: Base{SessionID: sid, Settings: s}}
	})
}

func (s *LocalSandbox) Create(ctx context.Context) error {
	port := os.Getenv("CODE_SERVER_PORT")
	if port == "" {
		port = "9000"
	}

	// water-ai local workspace logic
	workspacePath := fmt.Sprintf("/.water_ai/workspace/%s", s.SessionID)
	s.cmd = exec.CommandContext(ctx, "code-server", "--port", port, "--auth", "none", workspacePath)
	
	if err := s.cmd.Start(); err != nil {
		return err
	}

	s.HostURL = fmt.Sprintf("http://localhost:%d", s.Settings.SandboxConfig.ServicePort)
	return nil
}

func (s *LocalSandbox) Connect(ctx context.Context) error {
	s.HostURL = fmt.Sprintf("http://localhost:%d", s.Settings.SandboxConfig.ServicePort)
	return nil
}

func (s *LocalSandbox) Start(ctx context.Context) error   { return nil }
func (s *LocalSandbox) Stop(ctx context.Context) error    { return nil }
func (s *LocalSandbox) Cleanup(ctx context.Context) error { return nil }
func (s *LocalSandbox) ExposePort(port int) string        { return fmt.Sprintf("http://localhost:%d", port) }