package agent

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestTerminalToolRedactsStderrAndReportsBestEffortIsolation(t *testing.T) {
	result, err := (terminalExecTool{allowed: map[string]bool{"sh": true}}).Execute(context.Background(), ToolContext{Workspace: t.TempDir()}, map[string]any{
		"executable": "sh",
		"args":       []string{"-c", "printf 'api_key=supersecret123\\n' >&2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	value, ok := result.Value.(map[string]any)
	if !ok {
		t.Fatalf("result value = %#v", result.Value)
	}
	stderr, _ := value["stderr"].(string)
	if strings.Contains(stderr, "supersecret123") || !strings.Contains(stderr, "[REDACTED]") {
		t.Fatalf("stderr was not redacted: %q", stderr)
	}
	if value["execution_isolation"] != "best-effort-process-group" {
		t.Fatalf("execution isolation = %#v", value["execution_isolation"])
	}
}

func TestRunToolCommandKillsProcessGroupOnCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	command := exec.Command("sh", "-c", "sleep 10")
	configureToolProcess(command)
	started := time.Now()
	err := runToolCommand(ctx, command)
	if err == nil || !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("runToolCommand error = %v", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("cancellation took too long: %s", elapsed)
	}
}

func TestTerminalArgumentPolicyRejectsEscapeAndUnsupportedFlags(t *testing.T) {
	workspace := t.TempDir()
	if err := validateTerminalArguments("ls", []string{"."}, workspace); err != nil {
		t.Fatal(err)
	}
	if err := validateTerminalArguments("ls", []string{"../"}, workspace); err == nil {
		t.Fatal("expected ls path escape rejection")
	}
	if err := validateTerminalArguments("ls", []string{"--color=always"}, workspace); err == nil {
		t.Fatal("expected unsupported ls flag rejection")
	}
	if err := validateTerminalArguments("pwd", []string{"."}, workspace); err == nil {
		t.Fatal("expected pwd argument rejection")
	}
	if err := validateTerminalArguments("git", []string{"status", "--short"}, workspace); err == nil {
		t.Fatal("expected git argument rejection")
	}
}

func TestSandboxStrictModeFailsClosedWithoutDelegatedCgroup(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_SANDBOX_CGROUP_ROOT", t.TempDir())
	_, err := newSandboxControl("step_strict_test")
	if err == nil {
		t.Fatalf("strict sandbox error = %v", err)
	}
}

func TestSandboxRejectsUnknownMode(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_SANDBOX_MODE", "unsafe")
	_, err := (sandboxExecTool{}).Execute(context.Background(), ToolContext{Workspace: t.TempDir(), StepID: "step_mode_test"}, map[string]any{
		"language": "python",
		"code":     "print('must not run')",
	})
	if err == nil || !strings.Contains(err.Error(), "must be best-effort or strict") {
		t.Fatalf("unknown sandbox mode error = %v", err)
	}
}
