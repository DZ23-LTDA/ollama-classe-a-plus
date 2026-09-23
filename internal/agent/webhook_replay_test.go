package agent

import (
	"errors"
	"testing"
)

func TestWebhookReplayClaimPersistsAndRejectsReplay(t *testing.T) {
	root := t.TempDir()
	store, err := NewWebhookReplayStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Claim("sch_a", "evt_1"); err != nil {
		t.Fatal(err)
	}
	if err := store.Claim("sch_a", "evt_1"); !errors.Is(err, ErrWebhookReplay) {
		t.Fatalf("same-process replay error = %v", err)
	}
	reloaded, err := NewWebhookReplayStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := reloaded.Claim("sch_a", "evt_1"); !errors.Is(err, ErrWebhookReplay) {
		t.Fatalf("restart replay error = %v", err)
	}
	if err := reloaded.Claim("sch_b", "evt_1"); err != nil {
		t.Fatal(err)
	}
}
