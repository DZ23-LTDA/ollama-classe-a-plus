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
	manager := NewRemoteMCPManager()
	if err := manager.Register(RemoteMCPServerConfig{ID: "desktop", URL: server.URL, TokenEnv: "TEST_REMOTE_MCP_TOKEN", AllowedMethods: []string{"tools/list"}, TimeoutSeconds: 5}); err != nil {
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
