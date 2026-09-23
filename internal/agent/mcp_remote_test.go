package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestRemoteMCPCallRejectsCrossOrganizationServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"ok":true}}`))
	}))
	defer server.Close()
	manager := NewRemoteMCPManager()
	if err := manager.Register(RemoteMCPServerConfig{ID: "org-b", OrganizationID: "org_b", URL: server.URL, AllowedMethods: []string{"ping"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CallForOrganization(context.Background(), "org_a", "org-b", "ping", nil); !errors.Is(err, ErrPluginOrganizationScope) {
		t.Fatalf("cross-organization error = %v", err)
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
	if err := manager.Register(RemoteMCPServerConfig{ID: "blank", URL: "https://example.com/mcp", AllowedMethods: []string{" ", "\t"}}); err == nil {
		t.Fatal("expected blank remote MCP allowlist rejection")
	}
}

func TestRemoteMCPSSESupportsMultilineDataAndNotifications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: notification\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"progress\",\"params\":{}}\n\n"))
		_, _ = w.Write([]byte("event: response\ndata: {\"jsonrpc\":\"2.0\",\"id\":1,\ndata: \"result\":{\"ok\":true}}\n\n"))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}))
	defer server.Close()
	manager := NewRemoteMCPManager()
	if err := manager.Register(RemoteMCPServerConfig{ID: "multiline", URL: server.URL, AllowedMethods: []string{"ping"}, TimeoutSeconds: 3}); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Call(context.Background(), "multiline", "ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != `{"ok":true}` {
		t.Fatalf("result = %s", result)
	}
}

func TestRemoteMCPSSEReturnsCorrelatedErrorBeforeEOF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"jsonrpc\":\"2.0\",\"id\":1,\"error\":{\"code\":-32000,\"message\":\"denied\"}}\n\n"))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		<-time.After(2 * time.Second)
	}))
	defer server.Close()
	manager := NewRemoteMCPManager()
	if err := manager.Register(RemoteMCPServerConfig{ID: "error", URL: server.URL, AllowedMethods: []string{"ping"}, TimeoutSeconds: 5}); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	_, err := manager.Call(context.Background(), "error", "ping", nil)
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("err = %v, want correlated MCP error", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("correlated error waited for EOF: %s", elapsed)
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

func TestRemoteMCPRedirectsStaySameOrigin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"redirected":true}}`))
	}))
	defer server.Close()
	manager := NewRemoteMCPManager()
	if err := manager.Register(RemoteMCPServerConfig{ID: "redirect", URL: server.URL + "/redirect", AllowedMethods: []string{"ping"}}); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Call(context.Background(), "redirect", "ping", nil)
	if err != nil || !strings.Contains(string(result), "redirected") {
		t.Fatalf("redirect result=%s err=%v", result, err)
	}
}

func TestRemoteMCPDialRejectsPrivateActualAddress(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan error, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
		}
		accepted <- acceptErr
	}()
	conn, err := remoteMCPDialContext(context.Background(), "tcp", listener.Addr().String())
	if err == nil {
		_ = conn.Close()
		t.Fatal("expected private connected address rejection")
	}
	if !strings.Contains(err.Error(), "private") {
		t.Fatalf("unexpected dial error: %v", err)
	}
	select {
	case acceptErr := <-accepted:
		t.Fatalf("private destination was contacted: %v", acceptErr)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestRemoteMCPDialUsesApprovedIPWithoutSecondDNSLookup(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan error, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
		}
		accepted <- acceptErr
	}()
	oldLookup := lookupRemoteMCPIPs
	defer func() { lookupRemoteMCPIPs = oldLookup }()
	lookups := 0
	lookupRemoteMCPIPs = func(context.Context, string) ([]net.IP, error) {
		lookups++
		return nil, errors.New("second DNS lookup must not happen")
	}
	ctx := context.WithValue(context.Background(), remoteMCPLoopbackContextKey{}, true)
	ctx = context.WithValue(ctx, remoteMCPApprovedIPsContextKey{}, []net.IP{net.ParseIP("127.0.0.1")})
	conn, err := remoteMCPDialContext(ctx, "tcp", "mcp.example.test:"+strings.Split(listener.Addr().String(), ":")[1])
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()
	if lookups != 0 {
		t.Fatalf("DNS lookup count = %d, want zero after pinning", lookups)
	}
	if acceptErr := <-accepted; acceptErr != nil {
		t.Fatal(acceptErr)
	}
}
