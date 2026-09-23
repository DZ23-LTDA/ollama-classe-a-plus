package agent

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJSONStoreRollsBackMissionMutationWhenPersistenceFails(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONStore(root)
	if err != nil {
		t.Fatal(err)
	}
	original := Mission{ID: "mis_atomic", Version: 1, Objective: "original", State: MissionReady, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := store.PutMission(original); err != nil {
		t.Fatal(err)
	}
	store.root = filepath.Join(root, "missing-parent", "store")
	updated := original
	updated.Version = 2
	updated.Objective = "must not remain in memory"
	if err := store.PutMissionIfVersion(updated, original.Version); err == nil {
		t.Fatal("write unexpectedly succeeded")
	}
	got, err := store.GetMission(original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != original.Version || got.Objective != original.Objective {
		t.Fatalf("mission memory diverged: got=%+v original=%+v", got, original)
	}
	reloaded, err := NewJSONStore(root)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reloaded.GetMission(original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Version != original.Version || persisted.Objective != original.Objective {
		t.Fatalf("mission disk diverged: got=%+v original=%+v", persisted, original)
	}
	newMission := Mission{ID: "mis_new", Version: 1, Objective: "new", State: MissionReady}
	if err := store.PutMission(newMission); err == nil {
		t.Fatal("new mission write unexpectedly succeeded")
	}
	if _, err := store.GetMission(newMission.ID); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed new mission remained in memory: %v", err)
	}
}

func TestJSONStoreRollsBackEventsWhenPersistenceFails(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONStore(root)
	if err != nil {
		t.Fatal(err)
	}
	first := Event{ID: "evt_first", MissionID: "mis_events", Type: "first", CreatedAt: time.Now().UTC()}
	if err := store.AppendEvent(first); err != nil {
		t.Fatal(err)
	}
	store.root = filepath.Join(root, "missing-parent", "store")
	second := Event{ID: "evt_second", MissionID: first.MissionID, Type: "second", CreatedAt: time.Now().UTC()}
	if err := store.AppendEvent(second); err == nil {
		t.Fatal("event write unexpectedly succeeded")
	}
	events, err := store.ListEvents(first.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != first.ID {
		t.Fatalf("events memory diverged: %+v", events)
	}
	reloaded, err := NewJSONStore(root)
	if err != nil {
		t.Fatal(err)
	}
	events, err = reloaded.ListEvents(first.MissionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != first.ID {
		t.Fatalf("events disk diverged: %+v", events)
	}
}
