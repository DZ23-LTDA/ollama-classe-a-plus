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
