package agent

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestCompanyCycleIdempotencySurvivesRestart(t *testing.T) {
	root := t.TempDir()
	store, err := NewCompanyStore(filepath.Join(root, "companies"))
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Cycle Co"})
	if err != nil {
		t.Fatal(err)
	}
	input := CompanyCycle{Name: "Daily", Objective: "Review backlog", Frequency: "daily", IntervalSeconds: 86400}
	first, firstCycle, replayed, err := store.AddCycleWithIdempotency(company.ID, input, "cycle-key")
	if err != nil || replayed {
		t.Fatalf("first cycle = %+v replayed=%v err=%v", firstCycle, replayed, err)
	}
	if firstCycle.ID == "" || len(first.Cycles) != 1 {
		t.Fatalf("first cycle not persisted: %+v", first)
	}
	reloaded, err := NewCompanyStore(filepath.Join(root, "companies"))
	if err != nil {
		t.Fatal(err)
	}
	second, secondCycle, replayed, err := reloaded.AddCycleWithIdempotency(company.ID, input, "cycle-key")
	if err != nil || !replayed || secondCycle.ID != firstCycle.ID || len(second.Cycles) != 1 {
		t.Fatalf("replay = cycle=%+v replayed=%v company=%+v err=%v", secondCycle, replayed, second, err)
	}
	if _, _, _, err := reloaded.AddCycleWithIdempotency(company.ID, CompanyCycle{Name: "Different", Objective: "Review backlog", Frequency: "daily", IntervalSeconds: 86400}, "cycle-key"); !errors.Is(err, ErrCompanyIdempotencyConflict) {
		t.Fatalf("cycle conflict err=%v", err)
	}
}

func TestCompanyRemoveCycleAlsoRemovesItsIdempotencyRecord(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Rollback Co"})
	if err != nil {
		t.Fatal(err)
	}
	created, cycle, _, err := store.AddCycleWithIdempotency(company.ID, CompanyCycle{Name: "Daily", Objective: "Review", IntervalSeconds: 60}, "cleanup-key")
	if err != nil || len(created.Idempotency) != 1 {
		t.Fatalf("created=%+v cycle=%+v err=%v", created, cycle, err)
	}
	removed, err := store.RemoveCycle(company.ID, cycle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed.Cycles) != 0 || len(removed.Idempotency) != 0 {
		t.Fatalf("cycle cleanup incomplete: %+v", removed)
	}
}
