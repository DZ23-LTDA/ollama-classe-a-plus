package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAgentMCPRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mcp.json")
	if err := os.WriteFile(path, []byte(`[{
  "id": "example",
  "command": "sh",
  "allowed_methods": ["tools/list"],
  "unexpected": true
}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OLLAMA_AGENT_MCP", path)
	if _, err := loadAgentMCP(); err == nil {
		t.Fatal("expected unknown MCP field to be rejected")
	}
}

func TestLoadAgentConnectorsRejectsUnknownFieldsAndTrailingJSON(t *testing.T) {
	unknownPath := filepath.Join(t.TempDir(), "connectors-unknown.json")
	unknown := []byte(`[{"id":"example","provider":"Example","base_url":"https://api.example.com","operations":[{"name":"read","methods":["GET"],"path_prefixes":["/"]}],"unexpected":true}]`)
	if err := os.WriteFile(unknownPath, unknown, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OLLAMA_AGENT_CONNECTORS", unknownPath)
	if _, err := loadAgentConnectors(t.TempDir()); err == nil {
		t.Fatal("expected unknown connector field to be rejected")
	}

	trailingPath := filepath.Join(t.TempDir(), "connectors-trailing.json")
	if err := os.WriteFile(trailingPath, []byte(`[] {}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OLLAMA_AGENT_CONNECTORS", trailingPath)
	if _, err := loadAgentConnectors(t.TempDir()); err == nil {
		t.Fatal("expected trailing connector JSON to be rejected")
	}
}
