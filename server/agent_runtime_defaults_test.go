package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDefaultAgentRuntimeUsesDurableDefaults(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	for _, name := range []string{
		"OLLAMA_AGENT_ROOT",
		"OLLAMA_AGENT_STORE",
		"OLLAMA_AGENT_AUTH_STORE",
		"OLLAMA_AGENT_DATABASE_URL",
		"OLLAMA_AGENT_REDIS_URL",
		"OLLAMA_AGENT_CONNECTORS",
		"OLLAMA_AGENT_MCP",
		"OLLAMA_AGENT_REMOTE_MCP",
		"OLLAMA_AGENT_DEPLOYMENTS",
		"OLLAMA_AGENT_MEDIA_BASE_URL",
		"OLLAMA_AGENT_MEDIA_API_KEY",
		"OLLAMA_AGENT_OTLP_ENDPOINT",
	} {
		t.Setenv(name, "")
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	runtime, err := newDefaultAgentRuntime()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(runtime.WorkspaceRoot(), filepath.Clean(configDir)+string(os.PathSeparator)) {
		t.Fatalf("workspace default escaped XDG config directory: %q", runtime.WorkspaceRoot())
	}
	if filepath.Base(runtime.WorkspaceRoot()) != "workspaces" {
		t.Fatalf("workspace default = %q, want workspaces directory", runtime.WorkspaceRoot())
	}
	if filepath.Base(runtime.DataRoot()) != "data" {
		t.Fatalf("data default = %q, want data directory", runtime.DataRoot())
	}
	if _, err := newAgentAPI(runtime); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(runtime.DataRoot(), "auth")); err != nil {
		t.Fatalf("durable auth store was not created: %v", err)
	}
}
