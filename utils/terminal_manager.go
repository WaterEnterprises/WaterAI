// --- START OF FILE utils/terminal_manager.go ---
package utils

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
)

type TerminalSession struct {
	ID        string
	Cmd       *exec.Cmd
	Pty       *os.File
	Output    string
	History   []string
	CurrentDir string
	mu        sync.Mutex
	Alive     bool
}

type TerminalManager struct {
	Sessions map[string]*TerminalSession
	Shell    string
	Cwd      string
	mu       sync.Mutex
}

func NewTerminalManager(shell, cwd string) *TerminalManager {
	if shell == "" {
		shell = "/bin/bash"
	}
	return &TerminalManager{
		Sessions: make(map[string]*TerminalSession),
		Shell:    shell,
		Cwd:      cwd,
	}
}

func (tm *TerminalManager) CreateSession(id string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	c := exec.Command(tm.Shell)
	if tm.Cwd != "" {
		c.Dir = tm.Cwd
	}
	
	// Create PTY
	f, err := pty.Start(c)
	if err != nil {
		return err
	}

	session := &TerminalSession{
		ID:    id,
		Cmd:   c,
		Pty:   f,
		Alive: true,
	}

	// Output reader loop
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := f.Read(buf)
			if err != nil {
				if err != io.EOF {
					// Log error?
				}
				session.mu.Lock()
				session.Alive = false
				session.mu.Unlock()
				break
			}
			session.mu.Lock()
			session.Output += string(buf[:n])
			session.mu.Unlock()
		}
	}()

	// Setup basic prompt environment
	// Note: In a real PTY, we might write export PS1=... here
	time.Sleep(100 * time.Millisecond)
	f.Write([]byte("export PS1='[CMD_END]\\n'\n")) 
	
	tm.Sessions[id] = session
	return nil
}

func (tm *TerminalManager) Exec(sessionID, command string, timeoutSec int) SessionResult {
	tm.mu.Lock()
	sess, ok := tm.Sessions[sessionID]
	tm.mu.Unlock()

	if !ok {
		// Auto-create
		if err := tm.CreateSession(sessionID); err != nil {
			return SessionResult{Success: false, Output: err.Error()}
		}
		sess = tm.Sessions[sessionID]
	}

	sess.mu.Lock()
	if !sess.Alive {
		sess.mu.Unlock()
		return SessionResult{Success: false, Output: "Session is dead"}
	}
	// Clear buffer before command
	sess.Output = "" 
	sess.mu.Unlock()

	// Send command
	sess.Pty.Write([]byte(command + "\n"))
	sess.Pty.Write([]byte("echo [CMD_END]\n")) // Marker hack

	// Wait loop
	start := time.Now()
	for {
		if time.Since(start).Seconds() > float64(timeoutSec) {
			return SessionResult{Success: false, Output: "Timeout waiting for command output"}
		}
		
		sess.mu.Lock()
		out := sess.Output
		sess.mu.Unlock()

		if strings.Contains(out, "[CMD_END]") {
			// Clean up output
			clean := strings.ReplaceAll(out, "[CMD_END]", "")
			sess.mu.Lock()
			sess.History = append(sess.History, clean)
			sess.mu.Unlock()
			return SessionResult{Success: true, Output: clean}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (tm *TerminalManager) Kill(sessionID string) SessionResult {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	sess, ok := tm.Sessions[sessionID]
	if !ok {
		return SessionResult{Success: false, Output: "Session not found"}
	}
	
	if sess.Cmd != nil && sess.Cmd.Process != nil {
		sess.Cmd.Process.Kill()
	}
	sess.Pty.Close()
	delete(tm.Sessions, sessionID)
	
	return SessionResult{Success: true, Output: "Killed"}
}