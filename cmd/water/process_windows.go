//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func configureDetachedProcess(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// CREATE_NEW_PROCESS_GROUP = 0x200
	// DETACHED_PROCESS = 0x00000008
	cmd.SysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP
}