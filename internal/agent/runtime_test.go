package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRuntimePersistsAndRunsReadMission(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONStore(filepath.Join(root, ".store"))
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewRuntime(RuntimeConfig{Store: store, Planner: RulePlanner{}, WorkspaceRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{Objective: "inspecionar o workspace"})
	if err != nil {
		t.Fatal(err)
	}
	if mission.State != MissionReady {
		t.Fatalf("state = %s, want READY", mission.State)
	}
	if err := runtime.Run(context.Background(), mission.ID); err != nil {
		t.Fatal(err)
	}
	completed, err := runtime.GetMission(mission.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != MissionCompleted || completed.Plan[0].State != StepSucceeded {
		t.Fatalf("mission = %+v", completed)
	}
	reloaded, err := NewJSONStore(filepath.Join(root, ".store"))
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reloaded.GetMission(mission.ID)
	if err != nil || persisted.State != MissionCompleted {
		t.Fatalf("persisted = %+v err=%v", persisted, err)
	}
	events, err := runtime.Events(mission.ID)
	if err != nil || len(events) < 3 {
		t.Fatalf("events=%d err=%v", len(events), err)
	}
}

func TestRuntimeRequiresApprovalBeforeWritingAndBuildsArtifact(t *testing.T) {
	root := t.TempDir()
	runtime, err := NewRuntime(RuntimeConfig{Store: NewMemoryStore(), Planner: fixedPlanner{steps: []Step{{ID: "step_1", Kind: "workspace.write", Title: "write", Risk: RiskWrite, RequiresApproval: true, State: StepPending, Input: map[string]any{"path": "result.txt", "content": "hello"}}}}, WorkspaceRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{Objective: "escrever resultado"})
	if err != nil {
		t.Fatal(err)
	}
	if mission.State != MissionAwaitingApproval || len(mission.Approvals) != 1 {
		t.Fatalf("mission before approval = %+v", mission)
	}
	if err := runtime.Run(context.Background(), mission.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "result.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("write happened before approval: %v", err)
	}
	mission, err = runtime.DecideApproval(mission.ID, mission.Approvals[0].ID, true, "approved for test")
	if err != nil || mission.State != MissionReady {
		t.Fatalf("approval result = %+v err=%v", mission, err)
	}
	if err := runtime.Run(context.Background(), mission.ID); err != nil {
		t.Fatal(err)
	}
	completed, err := runtime.GetMission(mission.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != MissionCompleted || len(completed.Artifacts) != 1 || completed.Artifacts[0].SHA256 == "" {
		t.Fatalf("completed = %+v", completed)
	}
	data, err := os.ReadFile(filepath.Join(root, "result.txt"))
	if err != nil || string(data) != "hello" {
		t.Fatalf("artifact content=%q err=%v", data, err)
	}
}

func TestWorkspaceToolsRejectTraversal(t *testing.T) {
	runtime, err := NewRuntime(RuntimeConfig{Store: NewMemoryStore(), WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{Objective: "ler arquivo"})
	if err != nil {
		t.Fatal(err)
	}
	tool, ok := runtime.tools.Get("workspace.read")
	if !ok {
		t.Fatal("workspace.read not registered")
	}
	_, err = tool.Execute(context.Background(), ToolContext{MissionID: mission.ID, StepID: "step_1", Workspace: mission.Workspace}, map[string]any{"path": "../secret"})
	if err == nil || !strings.Contains(err.Error(), "escapes workspace") {
		t.Fatalf("error = %v, want traversal rejection", err)
	}
}

func TestWorkspaceToolsRejectSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlink unavailable on this platform: %v", err)
	}
	runtime, err := NewRuntime(RuntimeConfig{Store: NewMemoryStore(), WorkspaceRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{Objective: "ler arquivo"})
	if err != nil {
		t.Fatal(err)
	}
	tool, ok := runtime.tools.Get("workspace.read")
	if !ok {
		t.Fatal("workspace.read not registered")
	}
	if _, err := tool.Execute(context.Background(), ToolContext{MissionID: mission.ID, StepID: "step_1", Workspace: mission.Workspace}, map[string]any{"path": "linked/secret.txt"}); err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}

type fixedPlanner struct {
	steps []Step
}

func (p fixedPlanner) Plan(_ context.Context, _ Mission) ([]Step, error) {
	return append([]Step(nil), p.steps...), nil
}

func TestContextStorePersistsProjectAndMemory(t *testing.T) {
	root := t.TempDir()
	store, err := NewContextStore(filepath.Join(root, "context"))
	if err != nil {
		t.Fatal(err)
	}
	project, err := store.CreateProject("DZ23", filepath.Join(root, "workspace"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddMemory(Memory{ProjectID: project.ID, Kind: "decision", Content: "usar aprovação antes de publicar", Confidence: 1, Source: "test"}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewContextStore(filepath.Join(root, "context"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reloaded.GetProject(project.ID); err != nil {
		t.Fatal(err)
	}
	memories := reloaded.SearchMemories(project.ID, "aprovação", 10)
	if len(memories) != 1 || memories[0].Content == "" {
		t.Fatalf("memories = %+v", memories)
	}
}

func TestSandboxExecRunsIsolatedPython(t *testing.T) {
	root := t.TempDir()
	runtime, err := NewRuntime(RuntimeConfig{Store: NewMemoryStore(), WorkspaceRoot: root, Planner: fixedPlanner{steps: []Step{{ID: "step_1", Kind: "sandbox.exec", Title: "run", Risk: RiskWrite, RequiresApproval: true, State: StepPending, Input: map[string]any{"language": "python", "code": "print(2 + 2)"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{Objective: "executar código", Capabilities: []string{"sandbox:execute"}})
	if err != nil {
		t.Fatal(err)
	}
	mission, err = runtime.DecideApproval(mission.ID, mission.Approvals[0].ID, true, "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), mission.ID); err != nil {
		t.Fatal(err)
	}
	completed, err := runtime.GetMission(mission.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != MissionCompleted || !strings.Contains(completed.Plan[0].Result.(map[string]any)["stdout"].(string), "4") {
		t.Fatalf("completed = %+v", completed)
	}
}

func TestApprovalBindsActorOrganizationAndReason(t *testing.T) {
	runtime, err := NewRuntime(RuntimeConfig{Store: NewMemoryStore(), WorkspaceRoot: t.TempDir(), Planner: fixedPlanner{steps: []Step{{ID: "step_1", Kind: "workspace.write", Title: "write", Risk: RiskWrite, RequiresApproval: true, Input: map[string]any{"path": "approval.txt", "content": "ok"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{Objective: "aprovar operação", OrganizationID: "org_a"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.DecideApprovalForActor(mission.ID, mission.Approvals[0].ID, true, "approved", "user_a", "org_b"); err == nil {
		t.Fatal("expected organization mismatch")
	}
	if _, err := runtime.DecideApprovalForActor(mission.ID, mission.Approvals[0].ID, true, "", "user_a", "org_a"); err == nil {
		t.Fatal("expected empty reason rejection")
	}
	mission, err = runtime.DecideApprovalForActor(mission.ID, mission.Approvals[0].ID, true, "approved", "user_a", "org_a")
	if err != nil {
		t.Fatal(err)
	}
	approval := mission.Approvals[0]
	if approval.ActorID != "user_a" || approval.OrganizationID != "org_a" || approval.Nonce == "" || approval.Policy == "" || approval.ExpiresAt == nil {
		t.Fatalf("approval metadata = %+v", approval)
	}
}

func TestCapabilityPolicyDefaultsToLocalScopes(t *testing.T) {
	if !capabilityAllowed(ToolDescriptor{Name: "workspace.read", Scopes: []string{"workspace:read"}}, []string{"workspace:read", "workspace:write"}) {
		t.Fatal("local workspace read should be allowed by the default policy")
	}
	if capabilityAllowed(ToolDescriptor{Name: "browser.operator", Scopes: []string{"browser:navigate"}}, []string{"workspace:read", "workspace:write"}) {
		t.Fatal("browser capability must require explicit grant")
	}
	if !capabilityAllowed(ToolDescriptor{Name: "browser.operator", Scopes: []string{"browser:navigate"}}, []string{"browser:navigate"}) {
		t.Fatal("explicit browser capability should be accepted")
	}
	got := normalizeMissionCapabilities([]string{" browser:navigate ", "workspace:read", "browser:navigate"})
	if strings.Join(got, ",") != "browser:navigate,workspace:read" {
		t.Fatalf("normalized capabilities = %v", got)
	}
}

func TestScheduleClaimIsIdempotent(t *testing.T) {
	root := t.TempDir()
	store, err := NewContextStore(filepath.Join(root, "context"))
	if err != nil {
		t.Fatal(err)
	}
	schedule, err := store.CreateSchedule(Schedule{Objective: "verificar status", IntervalSeconds: 60})
	if err != nil {
		t.Fatal(err)
	}
	checkAt := time.Now().UTC().Add(2 * time.Minute)
	due := store.ClaimDueSchedules(checkAt)
	if len(due) != 1 || due[0].ID != schedule.ID {
		t.Fatalf("due = %+v", due)
	}
	if again := store.ClaimDueSchedules(checkAt); len(again) != 0 {
		t.Fatalf("schedule claimed twice: %+v", again)
	}
}

func TestBrowserOperatorNavigateAndSnapshot(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_BROWSER_ALLOW_PRIVATE", "1")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><head><title>DZ23</title></head><body><h1>Hello Browser</h1></body></html>"))
	}))
	defer server.Close()
	root := t.TempDir()
	tool := browserOperatorTool{}
	toolContext := ToolContext{MissionID: "browser-test", StepID: "step_1", Workspace: root}
	if _, err := tool.Execute(context.Background(), toolContext, map[string]any{"action": "navigate", "url": server.URL}); err != nil {
		t.Fatal(err)
	}
	result, err := tool.Execute(context.Background(), toolContext, map[string]any{"action": "snapshot"})
	if err != nil {
		t.Fatal(err)
	}
	content, ok := result.Value.(map[string]any)["content"].(string)
	if !ok || !strings.Contains(content, "Hello Browser") {
		t.Fatalf("browser result = %+v", result.Value)
	}
}
