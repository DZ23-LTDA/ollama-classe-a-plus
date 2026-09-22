//go:build windows

package agent

import "os/exec"

func configureMCPProcess(_ *exec.Cmd) {}

func terminateMCPProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
