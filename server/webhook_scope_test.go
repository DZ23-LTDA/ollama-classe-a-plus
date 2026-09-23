package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/agent"
)

func webhookTestContext(t *testing.T, organizationID, scheduleID, secret, idempotencyKey string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/agent/v1/webhooks/"+scheduleID, bytes.NewBufferString(`{"event":"created"}`))
	ctx.Request.Header.Set("X-Ollama-Agent-Secret", secret)
	if idempotencyKey != "" {
		ctx.Request.Header.Set("Idempotency-Key", idempotencyKey)
	}
	ctx.Params = gin.Params{{Key: "schedule_id", Value: scheduleID}}
	ctx.Set("agent.organization", agent.Organization{ID: organizationID})
	return ctx, recorder
}

func TestWebhookBindsScheduleOrganizationAndDeduplicates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("WEBHOOK_TEST_SECRET", "webhook-secret")
	contextStore, err := agent.NewContextStore("")
	if err != nil {
		t.Fatal(err)
	}
	schedule, err := contextStore.CreateSchedule(agent.Schedule{Objective: "process webhook", OrganizationID: "org_a", IntervalSeconds: 60, Enabled: true, WebhookSecretEnv: "WEBHOOK_TEST_SECRET"})
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := agent.NewRuntime(agent.RuntimeConfig{Context: contextStore, Planner: agent.RulePlanner{}, WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	api := &agentAPI{runtime: runtime, context: contextStore, authRequired: true}

	crossTenant, crossTenantRecorder := webhookTestContext(t, "org_b", schedule.ID, "webhook-secret", "evt-1")
	api.webhook(crossTenant)
	if crossTenantRecorder.Code != http.StatusForbidden {
		t.Fatalf("cross-tenant webhook status=%d body=%s", crossTenantRecorder.Code, crossTenantRecorder.Body.String())
	}

	missingKey, missingKeyRecorder := webhookTestContext(t, "org_a", schedule.ID, "webhook-secret", "")
	api.webhook(missingKey)
	if missingKeyRecorder.Code != http.StatusBadRequest {
		t.Fatalf("missing key status=%d body=%s", missingKeyRecorder.Code, missingKeyRecorder.Body.String())
	}

	first, firstRecorder := webhookTestContext(t, "org_a", schedule.ID, "webhook-secret", "evt-1")
	api.webhook(first)
	if firstRecorder.Code != http.StatusAccepted {
		t.Fatalf("first webhook status=%d body=%s", firstRecorder.Code, firstRecorder.Body.String())
	}

	replay, replayRecorder := webhookTestContext(t, "org_a", schedule.ID, "webhook-secret", "evt-1")
	api.webhook(replay)
	if replayRecorder.Code != http.StatusConflict {
		t.Fatalf("replay webhook status=%d body=%s", replayRecorder.Code, replayRecorder.Body.String())
	}

}
