package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/agent"
)

func newDeploymentApprovalTestContext(t *testing.T, method, path, body, builderID, provider string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: builderID}, {Key: "provider", Value: provider}}
	ctx.Set("agent.organization", agent.Organization{ID: "local"})
	ctx.Set("agent.user", agent.User{ID: "operator"})
	ctx.Set("agent.membership", agent.Membership{UserID: "operator", OrganizationID: "local", Role: agent.RoleOperator})
	return ctx, recorder
}

func TestDeployBuilderDoesNotTrustClientApprovedBoolean(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var requests atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"deployment-1","url":"http://127.0.0.1/site"}`))
	}))
	defer provider.Close()
	builder, err := agent.NewBuilderService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	project, err := builder.Create(context.Background(), agent.BuilderSpec{Name: "site", OrganizationID: "local", Kind: agent.BuilderWebsite})
	if err != nil {
		t.Fatal(err)
	}
	deployments := agent.NewDeploymentManager()
	if err := deployments.Register(agent.DeployConfig{ID: "self", Provider: "generic", BaseURL: provider.URL}); err != nil {
		t.Fatal(err)
	}
	runtime, err := agent.NewRuntime(agent.RuntimeConfig{WorkspaceRoot: t.TempDir(), Builder: builder, Deployments: deployments})
	if err != nil {
		t.Fatal(err)
	}
	api := &agentAPI{runtime: runtime}
	ctx, recorder := newDeploymentApprovalTestContext(t, http.MethodPost, "/builders/"+project.ID+"/deploy/self", `{"approved":true}`, project.ID, "self")
	api.deployBuilder(ctx)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("approved boolean status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if requests.Load() != 0 {
		t.Fatal("provider was called by client boolean")
	}
}

func TestDeploymentApprovalRequiresAdminAndNonceBeforeProviderCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var requests atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"deployment-1","url":"http://127.0.0.1/site"}`))
	}))
	defer provider.Close()
	builder, err := agent.NewBuilderService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	project, err := builder.Create(context.Background(), agent.BuilderSpec{Name: "site", OrganizationID: "local", Kind: agent.BuilderWebsite})
	if err != nil {
		t.Fatal(err)
	}
	deployments := agent.NewDeploymentManager()
	if err := deployments.Register(agent.DeployConfig{ID: "self", Provider: "generic", BaseURL: provider.URL}); err != nil {
		t.Fatal(err)
	}
	approvals, err := agent.NewDeploymentApprovalStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := agent.NewRuntime(agent.RuntimeConfig{WorkspaceRoot: t.TempDir(), Builder: builder, Deployments: deployments, DeploymentApprovals: approvals})
	if err != nil {
		t.Fatal(err)
	}
	api := &agentAPI{runtime: runtime, authRequired: true}

	approval, err := approvals.Request("local", project.ID, "self", "staging", "operator")
	if err != nil {
		t.Fatal(err)
	}
	operatorCtx, operatorRecorder := newDeploymentApprovalTestContext(t, http.MethodPost, "/builders/"+project.ID+"/deploy/self/approval/"+approval.ID, `{"approved":true,"nonce":"`+approval.Nonce+`"}`, project.ID, "self")
	operatorCtx.Params = append(operatorCtx.Params, gin.Param{Key: "approval_id", Value: approval.ID})
	api.decideDeploymentApproval(operatorCtx)
	if operatorRecorder.Code != http.StatusForbidden {
		t.Fatalf("operator approval status=%d body=%s", operatorRecorder.Code, operatorRecorder.Body.String())
	}
	if requests.Load() != 0 {
		t.Fatal("provider was called before admin approval")
	}

	adminCtx, adminRecorder := newDeploymentApprovalTestContext(t, http.MethodPost, "/builders/"+project.ID+"/deploy/self/approval/"+approval.ID, `{"approved":true,"nonce":"`+approval.Nonce+`"}`, project.ID, "self")
	adminCtx.Set("agent.membership", agent.Membership{UserID: "admin", OrganizationID: "local", Role: agent.RoleAdmin})
	adminCtx.Set("agent.user", agent.User{ID: "admin"})
	adminCtx.Params = append(adminCtx.Params, gin.Param{Key: "approval_id", Value: approval.ID})
	api.decideDeploymentApproval(adminCtx)
	if adminRecorder.Code != http.StatusOK {
		t.Fatalf("admin approval status=%d body=%s", adminRecorder.Code, adminRecorder.Body.String())
	}

	var approvalPayload agent.DeploymentApproval
	if err := json.Unmarshal(adminRecorder.Body.Bytes(), &approvalPayload); err != nil {
		t.Fatal(err)
	}
	deployCtx, deployRecorder := newDeploymentApprovalTestContext(t, http.MethodPost, "/builders/"+project.ID+"/deploy/self", `{"target":"staging","approval_id":"`+approval.ID+`","nonce":"`+approvalPayload.Nonce+`"}`, project.ID, "self")
	deployCtx.Set("agent.membership", agent.Membership{UserID: "operator", OrganizationID: "local", Role: agent.RoleOperator})
	api.deployBuilder(deployCtx)
	if deployRecorder.Code != http.StatusAccepted {
		t.Fatalf("deploy status=%d body=%s", deployRecorder.Code, deployRecorder.Body.String())
	}
	if requests.Load() != 1 {
		t.Fatalf("provider requests=%d, want 1", requests.Load())
	}

	replayCtx, replayRecorder := newDeploymentApprovalTestContext(t, http.MethodPost, "/builders/"+project.ID+"/deploy/self", `{"target":"staging","approval_id":"`+approval.ID+`","nonce":"`+approvalPayload.Nonce+`"}`, project.ID, "self")
	api.deployBuilder(replayCtx)
	if replayRecorder.Code != http.StatusConflict {
		t.Fatalf("replay status=%d body=%s", replayRecorder.Code, replayRecorder.Body.String())
	}
	if requests.Load() != 1 {
		t.Fatalf("provider requests after replay=%d, want 1", requests.Load())
	}
}
