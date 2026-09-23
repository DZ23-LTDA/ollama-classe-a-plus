//go:build !windows

package agent

import (
	"os/exec"
	"syscall"
)

func configureMCPProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminateMCPProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if cmd.SysProcAttr != nil && cmd.SysProcAttr.Setpgid {
		if processGroup, err := syscall.Getpgid(cmd.Process.Pid); err == nil && processGroup != syscall.Getpgrp() {
			return syscall.Kill(-processGroup, syscall.SIGKILL)
		}
	}
	return cmd.Process.Kill()
}
