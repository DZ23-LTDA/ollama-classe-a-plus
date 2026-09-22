package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemoteMCPCallJSONAndBearer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("x-consumer-api-key") != "composio-consumer" {
			t.Fatalf("x-consumer-api-key = %q", r.Header.Get("x-consumer-api-key"))
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request["method"] != "tools/list" {
			t.Fatalf("method = %#v", request["method"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"read_file"}]}}`))
	}))
	defer server.Close()
	t.Setenv("TEST_REMOTE_MCP_TOKEN", "test-token")
	t.Setenv("COMPOSIO_CONSUMER_KEY", "composio-consumer")
	manager := NewRemoteMCPManager()
	if err := manager.Register(RemoteMCPServerConfig{ID: "desktop", URL: server.URL, TokenEnv: "TEST_REMOTE_MCP_TOKEN", HeadersEnv: map[string]string{"x-consumer-api-key": "COMPOSIO_CONSUMER_KEY"}, AllowedMethods: []string{"tools/list"}, TimeoutSeconds: 5}); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Call(context.Background(), "desktop", "tools/list", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(result), "read_file") {
		t.Fatalf("result = %s", result)
	}
	if _, err := manager.Call(context.Background(), "desktop", "tools/call", nil); err == nil {
		t.Fatal("expected allowlist rejection")
	}
}

func TestRemoteMCPSSEAndURLPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"ok\":true}}\n\n"))
	}))
	defer server.Close()
	manager := NewRemoteMCPManager()
	if err := manager.Register(RemoteMCPServerConfig{ID: "local", URL: server.URL, AllowedMethods: []string{"ping"}}); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Call(context.Background(), "local", "ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != `{"ok":true}` {
		t.Fatalf("SSE result = %s", result)
	}
	if err := manager.Register(RemoteMCPServerConfig{ID: "external-http", URL: "http://example.com/mcp"}); err == nil {
		t.Fatal("expected external HTTP rejection")
	}
}

func TestRemoteMCPRequiresAllowlist(t *testing.T) {
	manager := NewRemoteMCPManager()
	if err := manager.Register(RemoteMCPServerConfig{ID: "empty", URL: "https://example.com/mcp"}); err == nil {
		t.Fatal("expected empty remote MCP allowlist rejection")
	}
}

func TestRemoteMCPRejectsInvalidHeaderEnvironment(t *testing.T) {
	manager := NewRemoteMCPManager()
	config := RemoteMCPServerConfig{
		ID:             "invalid-header",
		URL:            "https://example.com/mcp",
		AllowedMethods: []string{"tools/list"},
		HeadersEnv:     map[string]string{"X-Test": "INVALID-NAME"},
	}
	if err := manager.Register(config); err == nil {
		t.Fatal("expected invalid environment name rejection")
	}
	config.HeadersEnv = map[string]string{"Host": "VALID_NAME"}
	if err := manager.Register(config); err == nil {
		t.Fatal("expected restricted transport header rejection")
	}
}
