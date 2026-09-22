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
