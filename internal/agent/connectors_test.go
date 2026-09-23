package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestConnectorRequiresOrganizationScope(t *testing.T) {
	manager := NewConnectorManager()
	if _, _, err := manager.Call(context.Background(), "github", "profile", "GET", "/user", nil); err == nil {
		t.Fatal("expected direct connector call to require organization scope")
	}
	if _, _, err := manager.CallForOrganization(context.Background(), "", "github", "profile", "GET", "/user", nil); err == nil {
		t.Fatal("expected empty organization scope to be rejected")
	}
}

func TestConnectorListReportsCredentialStateWithoutTokenEnvironment(t *testing.T) {
	t.Setenv("CONNECTOR_STATUS_TOKEN", "configured-token")
	manager := NewConnectorManager()
	if err := manager.Register(ConnectorConfig{ID: "status", Provider: "status", BaseURL: "https://example.test", TokenEnv: "CONNECTOR_STATUS_TOKEN", Operations: []ConnectorOperation{{Name: "read", Methods: []string{"GET"}, PathPrefixes: []string{"/"}}}}); err != nil {
		t.Fatal(err)
	}
	listed := manager.List()
	if len(listed) != 1 || !listed[0].CredentialConfigured || listed[0].TokenEnv != "" {
		t.Fatalf("connector status = %+v", listed)
	}
}

func TestConnectorFailsClosedBeforeEgressWhenTokenEnvIsMissing(t *testing.T) {
	t.Setenv("CONNECTOR_MISSING_TOKEN", "")
	requests := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	manager := NewConnectorManager()
	manager.client = server.Client()
	if err := manager.Register(ConnectorConfig{ID: "missing-token", Provider: "test", BaseURL: server.URL, TokenEnv: "CONNECTOR_MISSING_TOKEN", Operations: []ConnectorOperation{{Name: "read", Methods: []string{"GET"}, PathPrefixes: []string{"/"}}}}); err != nil {
		t.Fatal(err)
	}
	_, _, err := manager.CallForOrganization(context.Background(), "org_test", "missing-token", "read", "GET", "/resource", nil)
	if err == nil || !strings.Contains(err.Error(), "credential") {
		t.Fatalf("missing credential error=%v", err)
	}
	if requests != 0 {
		t.Fatalf("connector made %d request(s) without credential", requests)
	}
}

func TestConnectorPathPrefixMatchesSegments(t *testing.T) {
	if !connectorPathMatches("/users/123", "/users") {
		t.Fatal("expected child path to match")
	}
	if connectorPathMatches("/users-privileged", "/users") {
		t.Fatal("must not match a different path segment")
	}
}

func TestConnectorEnforcesAllowedOrigins(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	op := ConnectorOperation{Name: "read", Methods: []string{"GET"}, PathPrefixes: []string{"/"}}

	blocked := NewConnectorManager()
	blocked.client = server.Client()
	if err := blocked.Register(ConnectorConfig{ID: "c", Provider: "test", BaseURL: server.URL, AllowedOrigins: []string{"https://not-allowed.example.com"}, Operations: []ConnectorOperation{op}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := blocked.CallForOrganization(context.Background(), "org_test", "c", "read", "GET", "/resource", nil); err == nil || !strings.Contains(err.Error(), "allowed_origins") {
		t.Fatalf("expected allowed_origins rejection, got %v", err)
	}

	allowed := NewConnectorManager()
	allowed.client = server.Client()
	if err := allowed.Register(ConnectorConfig{ID: "c", Provider: "test", BaseURL: server.URL, AllowedOrigins: []string{server.URL}, Operations: []ConnectorOperation{op}}); err != nil {
		t.Fatal(err)
	}
	if status, _, err := allowed.CallForOrganization(context.Background(), "org_test", "c", "read", "GET", "/resource", nil); err != nil || status != http.StatusOK {
		t.Fatalf("status=%d err=%v", status, err)
	}
}

func TestConnectorEgressBlocksRedirectsAndBoundsPayloads(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/large":
			_, _ = w.Write([]byte(strings.Repeat("x", 2<<20+1)))
		default:
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}))
	defer server.Close()
	manager := NewConnectorManager()
	manager.client = server.Client()
	if err := manager.Register(ConnectorConfig{ID: "egress", Provider: "test", BaseURL: server.URL, Operations: []ConnectorOperation{{Name: "read", Methods: []string{"GET", "POST"}, PathPrefixes: []string{"/"}}}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.CallForOrganization(context.Background(), "org_test", "egress", "read", "GET", "/redirect", nil); err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("expected redirect rejection, got %v", err)
	}
	if _, _, err := manager.CallForOrganization(context.Background(), "org_test", "egress", "read", "GET", "/large", nil); err == nil || !strings.Contains(err.Error(), "payload") {
		t.Fatalf("expected response limit rejection, got %v", err)
	}
	if _, _, err := manager.CallForOrganization(context.Background(), "org_test", "egress", "read", "POST", "/final", []byte(strings.Repeat("x", 1<<20+1))); err == nil || !strings.Contains(err.Error(), "payload") {
		t.Fatalf("expected request limit rejection, got %v", err)
	}
}
