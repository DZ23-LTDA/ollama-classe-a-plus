package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConnectorLifecycleDisablesCalls(t *testing.T) {
	manager := NewConnectorManager()
	if err := manager.Register(ConnectorConfig{ID: "c", Provider: "test", BaseURL: "https://example.test", Operations: []ConnectorOperation{{Name: "read", Methods: []string{"GET"}, PathPrefixes: []string{"/"}}}}); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetEnabled("c", false); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.CallForOrganization(context.Background(), "org_test", "c", "read", "GET", "/", nil); !errors.Is(err, ErrConnectorDisabled) {
		t.Fatalf("expected disabled connector, got %v", err)
	}
	if err := manager.SetEnabled("c", true); err != nil {
		t.Fatal(err)
	}
	if err := manager.Remove("c"); err != nil {
		t.Fatal(err)
	}
}

func TestMCPAndSkillsLifecycle(t *testing.T) {
	manager := NewMCPManager()
	if err := manager.Register(MCPServerConfig{ID: "echo", Command: os.Args[0], Args: []string{"-test.run=TestMCPHelperProcess"}, AllowedMethods: []string{"echo"}}); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetEnabled("echo", false); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Call(context.Background(), "echo", "echo", map[string]any{}); err == nil {
		t.Fatal("expected disabled MCP error")
	}
	if err := manager.Remove("echo"); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(`{"id":"skill","version":"1","description":"test","trusted":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := NewContextStore(filepath.Join(t.TempDir(), "context"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.LoadSkills(dir, true); err != nil {
		t.Fatal(err)
	}
	if len(store.Skills()) != 1 || !store.Skills()[0].Enabled || store.Skills()[0].Trusted {
		t.Fatalf("skills=%+v", store.Skills())
	}
	if err := store.SetSkillEnabled("skill", false); err != nil {
		t.Fatal(err)
	}
	if store.Skills()[0].Enabled {
		t.Fatal("skill should be disabled")
	}
	if err := store.RemoveSkill("skill"); err != nil {
		t.Fatal(err)
	}
}
