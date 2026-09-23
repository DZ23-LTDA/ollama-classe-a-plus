package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/agent"
)

// tenant seeds a user + organization and returns a valid bearer token for it.
func tenant(t *testing.T, auth *agent.AuthStore, email string) string {
	t.Helper()
	user, err := auth.CreateUser(email, email)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	org, _, err := auth.CreateOrganization(email+"-org", user)
	if err != nil {
		t.Fatalf("create org: %v", err)
	}
	token, _, err := auth.IssueToken(user.ID, org.ID, time.Hour)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return token
}

// TestAgentProjectTenantIsolation proves that, with authentication required, a
// project created by one organization is invisible to another (no cross-tenant
// IDOR): the owner reads it, a different tenant gets 404.
func TestAgentProjectTenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rt, err := agent.NewRuntime(agent.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("runtime: %v", err)
	}
	auth, err := agent.NewAuthStore("")
	if err != nil {
		t.Fatalf("auth store: %v", err)
	}
	api := &agentAPI{runtime: rt, context: rt.Context(), auth: auth, authRequired: true, samlServices: map[string]*agent.SAMLService{}}
	engine := gin.New()
	api.register(engine)

	tokenA := tenant(t, auth, "a@example.com")
	tokenB := tenant(t, auth, "b@example.com")

	do := func(method, path, token string, body []byte) *httptest.ResponseRecorder {
		var reader *bytes.Reader
		if body == nil {
			reader = bytes.NewReader(nil)
		} else {
			reader = bytes.NewReader(body)
		}
		req := httptest.NewRequest(method, path, reader)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		return rec
	}

	// Org A creates a project.
	created := do(http.MethodPost, "/api/agent/v1/projects", tokenA, []byte(`{"name":"tenant-a-project"}`))
	if created.Code != http.StatusCreated {
		t.Fatalf("create project by A: status=%d body=%s", created.Code, created.Body.String())
	}
	var project struct {
		ID             string `json:"id"`
		OrganizationID string `json:"organization_id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &project); err != nil {
		t.Fatalf("decode project: %v", err)
	}
	if project.ID == "" {
		t.Fatal("created project has no id")
	}

	// Org A can read its own project.
	if got := do(http.MethodGet, "/api/agent/v1/projects/"+project.ID, tokenA, nil); got.Code != http.StatusOK {
		t.Fatalf("owner read: status=%d body=%s", got.Code, got.Body.String())
	}

	// Org B must NOT see org A's project (cross-tenant read is hidden as 404).
	if got := do(http.MethodGet, "/api/agent/v1/projects/"+project.ID, tokenB, nil); got.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant read should be 404, got status=%d body=%s", got.Code, got.Body.String())
	}

	// Org B must NOT be able to write memories into org A's project.
	if got := do(http.MethodPost, "/api/agent/v1/projects/"+project.ID+"/memories", tokenB, []byte(`{"kind":"note","content":"x","confidence":1}`)); got.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant memory write should be 404, got status=%d body=%s", got.Code, got.Body.String())
	}
}
