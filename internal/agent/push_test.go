package agent

import (
	"os"
	"testing"
)

func TestPushServiceRegisterRollsBackOnPersistenceFailure(t *testing.T) {
	root := t.TempDir()
	service, err := NewPushService(root, "http://127.0.0.1:43123")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Register("push-token", "android", "user-a", "org-a"); err == nil {
		t.Fatal("register unexpectedly succeeded without a writable root")
	}
	if got := service.ListOrganization("org-a"); len(got) != 0 {
		t.Fatalf("subscription remained in memory after persistence failure: %+v", got)
	}
}
