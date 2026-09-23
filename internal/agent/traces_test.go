package agent

import (
	"os"
	"testing"
)

func TestTraceStoreOrganizationScope(t *testing.T) {
	store, err := NewTraceStore("")
	if err != nil {
		t.Fatal(err)
	}
	span := store.StartForOrganization("org-a", "tr_scope", "", "tenant span", nil)
	span.End("ok", nil)
	if got := store.ListForOrganization("org-b", "", 100); len(got) != 0 {
		t.Fatalf("cross-tenant trace list returned %+v", got)
	}
	if got := store.ListForOrganization("org-a", "tr_scope", 100); len(got) != 1 || got[0].OrganizationID != "org-a" {
		t.Fatalf("same-tenant trace list=%+v", got)
	}
}

func TestTraceStoreRollsBackSpanWhenPersistenceFails(t *testing.T) {
	root := t.TempDir()
	store, err := NewTraceStore(root)
	if err != nil {
		t.Fatal(err)
	}
	span := store.StartForOrganization("org-a", "tr_rollback", "", "operation", nil)
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	span.End("error", os.ErrPermission)
	spans := store.ListForOrganization("org-a", "tr_rollback", 10)
	if len(spans) != 1 || spans[0].Status != "running" || spans[0].EndAt != nil {
		t.Fatalf("span changed after persistence failure: %+v", spans)
	}
}
