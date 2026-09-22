package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMCPManagerCallsAllowlistedMethod(t *testing.T) {
	t.Setenv("GO_WANT_MCP_HELPER_PROCESS", "1")
	manager := NewMCPManager()
	if err := manager.Register(MCPServerConfig{ID: "echo", Command: os.Args[0], Args: []string{"-test.run=TestMCPHelperProcess"}, AllowedMethods: []string{"echo"}, EnvironmentVars: []string{"GO_WANT_MCP_HELPER_PROCESS"}, TimeoutSeconds: 5}); err != nil {
		t.Fatal(err)
	}
	defer manager.StopAll()
	result, err := manager.Call(context.Background(), "echo", "echo", map[string]any{"value": "hello"})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil || payload["value"] != "hello" {
		t.Fatalf("result=%s err=%v", result, err)
	}
	if _, err := manager.Call(context.Background(), "echo", "tools/list", nil); err == nil || !strings.Contains(err.Error(), "not allowlisted") {
		t.Fatalf("unexpected allowlist result: %v", err)
	}
}

func TestMCPNotificationsDoNotBreakResponseCorrelation(t *testing.T) {
	t.Setenv("GO_WANT_MCP_HELPER_PROCESS", "1")
	manager := NewMCPManager()
	if err := manager.Register(MCPServerConfig{ID: "notify", Command: os.Args[0], Args: []string{"-test.run=TestMCPHelperProcess"}, AllowedMethods: []string{"notify"}, EnvironmentVars: []string{"GO_WANT_MCP_HELPER_PROCESS"}, TimeoutSeconds: 5}); err != nil {
		t.Fatal(err)
	}
	defer manager.StopAll()
	result, err := manager.Call(context.Background(), "notify", "notify", map[string]any{"value": "after-notification"})
	if err != nil || !strings.Contains(string(result), "after-notification") {
		t.Fatalf("notification result=%s err=%v", result, err)
	}
}

func TestMCPHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_MCP_HELPER_PROCESS") != "1" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var request struct {
			ID     int64          `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			fmt.Printf("{\"jsonrpc\":\"2.0\",\"id\":0,\"error\":{\"message\":%q}}\n", err.Error())
			continue
		}
		if request.Method == "sleep" {
			time.Sleep(10 * time.Second)
			continue
		}
		if request.Method == "notify" {
			fmt.Printf("{\"jsonrpc\":\"2.0\",\"method\":\"progress\",\"params\":{\"status\":\"working\"}}\n")
		}
		value := request.Params["value"]
		fmt.Printf("{\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"value\":%q}}\n", request.ID, value)
	}
	os.Exit(0)
}

func TestMCPRequiresAllowlist(t *testing.T) {
	manager := NewMCPManager()
	if err := manager.Register(MCPServerConfig{ID: "empty", Command: os.Args[0]}); err == nil {
		t.Fatal("expected empty MCP allowlist rejection")
	}
}

func TestMCPRequiresAbsoluteExecutableAndUsesPrivateWorkspace(t *testing.T) {
	manager := NewMCPManager()
	if err := manager.Register(MCPServerConfig{ID: "relative", Command: "sh", AllowedMethods: []string{"echo"}}); err == nil {
		t.Fatal("expected absolute command rejection")
	}
	link := t.TempDir() + "/mcp-link"
	if err := os.Symlink(os.Args[0], link); err == nil {
		if err := manager.Register(MCPServerConfig{ID: "symlink", Command: link, AllowedMethods: []string{"echo"}}); err == nil {
			t.Fatal("expected symlink executable rejection")
		}
	}
	t.Setenv("GO_WANT_MCP_HELPER_PROCESS", "1")
	if err := manager.Register(MCPServerConfig{ID: "workspace", Command: os.Args[0], Args: []string{"-test.run=TestMCPHelperProcess"}, AllowedMethods: []string{"echo"}, EnvironmentVars: []string{"GO_WANT_MCP_HELPER_PROCESS"}}); err != nil {
		t.Fatal(err)
	}
	config := manager.List()[0]
	if config.WorkingDirectory == "" {
		t.Fatal("expected explicit MCP working directory")
	}
	if _, err := os.Stat(config.WorkingDirectory); err != nil {
		t.Fatalf("working directory missing: %v", err)
	}
	manager.StopAll()
	if _, err := os.Stat(config.WorkingDirectory); !os.IsNotExist(err) {
		t.Fatalf("temporary MCP workspace was not removed: %v", err)
	}
}

func TestMCPPayloadLimitAndCancellationRestart(t *testing.T) {
	t.Setenv("GO_WANT_MCP_HELPER_PROCESS", "1")
	manager := NewMCPManager()
	if err := manager.Register(MCPServerConfig{ID: "limits", Command: os.Args[0], Args: []string{"-test.run=TestMCPHelperProcess"}, AllowedMethods: []string{"echo", "sleep"}, EnvironmentVars: []string{"GO_WANT_MCP_HELPER_PROCESS"}, TimeoutSeconds: 1}); err != nil {
		t.Fatal(err)
	}
	defer manager.StopAll()
	if _, err := manager.Call(context.Background(), "limits", "echo", map[string]any{"value": strings.Repeat("x", mcpMaxMessageBytes)}); err == nil || !strings.Contains(err.Error(), "payload limit") {
		t.Fatalf("expected request payload limit, got %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := manager.Call(ctx, "limits", "sleep", nil); err == nil {
		t.Fatal("expected cancelled MCP call")
	}
	result, err := manager.Call(context.Background(), "limits", "echo", map[string]any{"value": "restarted"})
	if err != nil || !strings.Contains(string(result), "restarted") {
		t.Fatalf("restart result=%s err=%v", result, err)
	}
}
