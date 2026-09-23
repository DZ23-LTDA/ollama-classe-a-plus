package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testEmbedder map[string][]float32

func (e testEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	return e[text], nil
}

func TestSemanticMemorySearchRanksByCosineSimilarity(t *testing.T) {
	store, err := NewContextStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.SetEmbedder(testEmbedder{
		"alpha": {1, 0},
		"beta":  {0.8, 0.2},
		"query": {1, 0},
	})
	if _, err := store.AddMemoryContext(context.Background(), Memory{ProjectID: "project", Kind: "note", Content: "alpha", Confidence: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddMemoryContext(context.Background(), Memory{ProjectID: "project", Kind: "note", Content: "beta", Confidence: 1}); err != nil {
		t.Fatal(err)
	}
	memories, err := store.SearchMemoriesContext(context.Background(), "project", "query", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(memories) != 2 || memories[0].Content != "alpha" {
		t.Fatalf("memories = %+v", memories)
	}
}

func TestProjectAndScheduleCRUDPersistsAndDeletes(t *testing.T) {
	root := t.TempDir()
	store, err := NewContextStore(root)
	if err != nil {
		t.Fatal(err)
	}
	project, err := store.CreateProject("Workspace", "", "org_test")
	if err != nil {
		t.Fatal(err)
	}
	if got := store.ListProjects(); len(got) != 1 || got[0].OrganizationID != "org_test" {
		t.Fatalf("projects = %+v", got)
	}
	updatedProject, err := store.UpdateProject(project.ID, "Workspace atualizado", "")
	if err != nil || updatedProject.Name != "Workspace atualizado" {
		t.Fatalf("updated project = %+v, err = %v", updatedProject, err)
	}
	schedule, err := store.CreateSchedule(Schedule{Objective: "verificar", IntervalSeconds: 60, OrganizationID: "org_test"})
	if err != nil {
		t.Fatal(err)
	}
	if got := store.ListSchedulesForOrganization("org_test"); len(got) != 1 || got[0].ID != schedule.ID {
		t.Fatalf("schedules = %+v", got)
	}
	if _, err := store.UpdateSchedule(schedule.ID, Schedule{Objective: "verificar atualizado", IntervalSeconds: 120, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteSchedule(schedule.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteProject(project.ID); err != nil {
		t.Fatal(err)
	}
	if len(store.ListProjects()) != 0 || len(store.ListSchedules()) != 0 {
		t.Fatalf("store was not deleted: projects=%v schedules=%v", store.ListProjects(), store.ListSchedules())
	}
}

func TestProjectRootCreateAndUpdateStayInsideRuntimeWorkspace(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	store, err := NewContextStore("")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetWorkspaceRoot(workspace); err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(workspace, "project")
	if err := os.MkdirAll(inside, 0o700); err != nil {
		t.Fatal(err)
	}
	project, err := store.CreateProject("Workspace", inside)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateProject("Outside", outside); err == nil || !strings.Contains(err.Error(), "outside the runtime workspace") {
		t.Fatalf("outside create error = %v", err)
	}
	if _, err := store.UpdateProject(project.ID, "Outside", outside); err == nil || !strings.Contains(err.Error(), "outside the runtime workspace") {
		t.Fatalf("outside update error = %v", err)
	}
	current, err := store.GetProject(project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Root != project.Root {
		t.Fatalf("project root changed after rejected update: %q", current.Root)
	}
}
