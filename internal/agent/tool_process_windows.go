//go:build windows

package agent

import "os/exec"

func configureToolProcess(_ *exec.Cmd) {}

func terminateToolProcess(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
