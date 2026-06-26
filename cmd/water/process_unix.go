//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

func configureDetachedProcess(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Setsid creates a new session, detaching from TTY
	cmd.SysProcAttr.Setsid = true
}