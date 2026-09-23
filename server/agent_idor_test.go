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

	// Collaboration endpoints are project-scoped: org B cannot read or write org
	// A's collaboration surface.
	if got := do(http.MethodGet, "/api/agent/v1/collab/"+project.ID, tokenB, nil); got.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant collab read should be 404, got status=%d body=%s", got.Code, got.Body.String())
	}
	if got := do(http.MethodPost, "/api/agent/v1/collab/"+project.ID+"/comments", tokenB, []byte(`{"body":"intruso"}`)); got.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant collab comment should be 404, got status=%d body=%s", got.Code, got.Body.String())
	}
	// The owner can read its own collaboration snapshot.
	if got := do(http.MethodGet, "/api/agent/v1/collab/"+project.ID, tokenA, nil); got.Code != http.StatusOK {
		t.Fatalf("owner collab read: status=%d body=%s", got.Code, got.Body.String())
	}
}

// TestAgentTracesRequireOwningMission proves that the wildcard traces endpoint no
// longer dumps every tenant's spans: it requires a trace_id whose mission belongs
// to the caller's organization.
func TestAgentTracesRequireOwningMission(t *testing.T) {
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

	get := func(path, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		return rec
	}

	// Org A creates a mission (its trace id is "tr_<missionID>").
	created := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/agent/v1/missions", bytes.NewReader([]byte(`{"objective":"inspecionar o workspace"}`)))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(created, req)
	if created.Code != http.StatusCreated {
		t.Fatalf("create mission: status=%d body=%s", created.Code, created.Body.String())
	}
	var mission struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &mission); err != nil || mission.ID == "" {
		t.Fatalf("decode mission: %v", err)
	}

	// Empty trace_id is rejected (no more full dump).
	if got := get("/api/agent/v1/traces", tokenA); got.Code != http.StatusBadRequest {
		t.Fatalf("empty trace_id should be 400, got status=%d body=%s", got.Code, got.Body.String())
	}
	// Org B cannot read org A's mission spans by trace_id (blocked with a client
	// error — never 200 with data).
	if got := get("/api/agent/v1/traces?trace_id=tr_"+mission.ID, tokenB); got.Code == http.StatusOK {
		t.Fatalf("cross-tenant trace read must be blocked, got 200 body=%s", got.Body.String())
	}
	// The owner can.
	if got := get("/api/agent/v1/traces?trace_id=tr_"+mission.ID, tokenA); got.Code != http.StatusOK {
		t.Fatalf("owner trace read: status=%d body=%s", got.Code, got.Body.String())
	}
}
