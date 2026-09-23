package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContextStoreCreateScheduleRollsBackOnPersistenceFailure(t *testing.T) {
	root := t.TempDir()
	store, err := NewContextStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(root, "schedules")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "schedules"), []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateSchedule(Schedule{Objective: "will fail", IntervalSeconds: 60}); err == nil {
		t.Fatal("CreateSchedule unexpectedly succeeded")
	}
	if got := store.ListSchedules(); len(got) != 0 {
		t.Fatalf("failed schedule remained in memory: %+v", got)
	}
}
