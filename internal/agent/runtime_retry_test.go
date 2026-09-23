package agent

import (
	"context"
	"errors"
	"testing"
)

// flakyTool fails its first (failures) executions, then succeeds. It records how
// many times Execute was actually invoked so the retry path can be asserted.
type flakyTool struct {
	failures int
	calls    *int
}

func (flakyTool) Descriptor() ToolDescriptor {
	return ToolDescriptor{Name: "workspace.list", Version: "1", Description: "flaky test tool", Risk: RiskRead}
}

func (t flakyTool) Execute(_ context.Context, _ ToolContext, _ map[string]any) (ToolResult, error) {
	*t.calls++
	if *t.calls <= t.failures {
		return ToolResult{}, errors.New("transient failure")
	}
	return ToolResult{Value: map[string]any{"ok": true}}, nil
}

func newFlakyRuntime(t *testing.T, failures int, calls *int) *Runtime {
	t.Helper()
	tools := NewRegistry()
	tools.Register(flakyTool{failures: failures, calls: calls})
	step := Step{ID: "step_1", Kind: "workspace.list", Title: "flaky", Risk: RiskRead, State: StepPending}
	runtime, err := NewRuntime(RuntimeConfig{
		Store:         NewMemoryStore(),
		WorkspaceRoot: t.TempDir(),
		Tools:         tools,
		Planner:       fixedPlanner{steps: []Step{step}},
	})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	return runtime
}

// Regression for the retry bug: a transient failure on the first attempt must be
// retried (the step re-executes) and the mission must complete — not be marked
// COMPLETED while the failed step is silently left pending.
func TestRuntimeRetriesTransientStepThenCompletes(t *testing.T) {
	calls := 0
	runtime := newFlakyRuntime(t, 1, &calls)
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{Objective: "flaky mission"})
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}
	if err := runtime.Run(context.Background(), mission.ID); err != nil {
		t.Fatalf("run: %v", err)
	}
	done, err := runtime.store.GetMission(mission.ID)
	if err != nil {
		t.Fatalf("get mission: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected the step to execute twice (1 fail + 1 retry), got %d", calls)
	}
	if done.State != MissionCompleted {
		t.Fatalf("expected mission COMPLETED, got %s", done.State)
	}
	if done.Plan[0].State != StepSucceeded {
		t.Fatalf("expected step SUCCEEDED, got %s", done.Plan[0].State)
	}
}

// A step that keeps failing must end the mission FAILED, never COMPLETED.
func TestRuntimePersistentFailureDoesNotFalselyComplete(t *testing.T) {
	calls := 0
	runtime := newFlakyRuntime(t, 5, &calls)
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{Objective: "always fails"})
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}
	_ = runtime.Run(context.Background(), mission.ID)
	done, err := runtime.store.GetMission(mission.ID)
	if err != nil {
		t.Fatalf("get mission: %v", err)
	}
	if done.State == MissionCompleted {
		t.Fatalf("mission must not be COMPLETED when a step never succeeds")
	}
	if done.State != MissionFailed {
		t.Fatalf("expected mission FAILED, got %s", done.State)
	}
}
