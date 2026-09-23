package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConnectorEnforcesAllowedOrigins(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	op := ConnectorOperation{Name: "profile", Methods: []string{"GET"}, PathPrefixes: []string{"/user"}}

	// A non-matching allowed origin blocks the call before any request is made.
	blocked := NewConnectorManager()
	blocked.client = server.Client()
	if err := blocked.Register(ConnectorConfig{ID: "c", Provider: "p", BaseURL: server.URL, AllowedOrigins: []string{"https://not-allowed.example.com"}, Operations: []ConnectorOperation{op}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := blocked.Call(context.Background(), "c", "profile", "GET", "/user", nil); err == nil || !strings.Contains(err.Error(), "allowed_origins") {
		t.Fatalf("expected allowed_origins rejection, got %v", err)
	}

	// A matching allowed origin lets the call proceed.
	allowed := NewConnectorManager()
	allowed.client = server.Client()
	if err := allowed.Register(ConnectorConfig{ID: "c", Provider: "p", BaseURL: server.URL, AllowedOrigins: []string{server.URL}, Operations: []ConnectorOperation{op}}); err != nil {
		t.Fatal(err)
	}
	if status, _, err := allowed.Call(context.Background(), "c", "profile", "GET", "/user", nil); err != nil || status != http.StatusOK {
		t.Fatalf("status=%d err=%v", status, err)
	}
}

func TestConnectorUsesTenantOAuthCredential(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_CREDENTIAL_KEY", "connector-test-key")
	store, err := NewAuthStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.CreateUser("connector@example.com", "Connector User")
	if err != nil {
		t.Fatal(err)
	}
	organization, _, err := store.CreateOrganization("Connector Org", user)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Authorize(user.ID, organization.ID, "write"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.StoreOAuthCredential("github", user.ID, organization.ID, map[string]any{"access_token": "tenant-token", "expires_in": float64(3600)}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tenant-token" {
			t.Errorf("authorization=%q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	manager := NewConnectorManager()
	manager.client = server.Client()
	manager.SetOAuthStore(store)
	if err := manager.Register(ConnectorConfig{ID: "github", Provider: "github", BaseURL: server.URL, OAuthProvider: "github", Operations: []ConnectorOperation{{Name: "profile", Methods: []string{"GET"}, PathPrefixes: []string{"/user"}}}}); err != nil {
		t.Fatal(err)
	}
	status, response, err := manager.CallForOrganization(context.Background(), organization.ID, "github", "profile", "GET", "/user", nil)
	if err != nil || status != http.StatusOK || response != `{"ok":true}` {
		t.Fatalf("status=%d response=%q err=%v", status, response, err)
	}
}
