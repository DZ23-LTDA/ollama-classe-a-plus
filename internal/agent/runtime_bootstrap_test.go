package agent

import (
	"context"
	"strings"
	"testing"
)

func TestCreateMissionProviderResolutionFailurePersistsFailedState(t *testing.T) {
	store := NewMemoryStore()
	runtime, err := NewRuntime(RuntimeConfig{Store: store, WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	_, err = runtime.CreateMission(context.Background(), CreateMissionRequest{
		Objective: "executar provider ausente",
		Provider:  "codex",
		Model:     "codex-mini",
	})
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("err = %v, want explicit provider configuration error", err)
	}
	missions, err := store.ListMissions()
	if err != nil {
		t.Fatal(err)
	}
	if len(missions) != 1 {
		t.Fatalf("stored missions = %d, want 1 audit record", len(missions))
	}
	if missions[0].State != MissionFailed {
		t.Fatalf("mission state = %s, want FAILED", missions[0].State)
	}
	if !strings.Contains(missions[0].LastError, "not configured") {
		t.Fatalf("last error = %q", missions[0].LastError)
	}
	events, err := store.ListEvents(missions[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[1].Type != "mission.failed" {
		t.Fatalf("events = %+v, want created and mission.failed", events)
	}
}

func TestUnconfiguredPlannerFailsClosed(t *testing.T) {
	_, err := (UnconfiguredPlanner{}).Plan(context.Background(), Mission{Objective: "listar"})
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("err = %v, want unconfigured planner failure", err)
	}
}

func TestRuntimeUsesLocalModelResolver(t *testing.T) {
	resolver := &plannerResolverStub{planner: RulePlanner{}}
	runtime, err := NewRuntime(RuntimeConfig{
		Store:           NewMemoryStore(),
		PlannerResolver: resolver,
		WorkspaceRoot:   t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	mission, err := runtime.CreateMission(context.Background(), CreateMissionRequest{
		Objective: "listar o workspace",
		Provider:  "ollama-local",
		Model:     "qwen3-coder:latest",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resolver.provider != "ollama-local" || resolver.model != "qwen3-coder:latest" {
		t.Fatalf("resolver request = %q/%q", resolver.provider, resolver.model)
	}
	if mission.Model != "qwen3-coder:latest" {
		t.Fatalf("mission model = %q", mission.Model)
	}
}
