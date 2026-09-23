package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/grok"
)

func TestGrokResponsesRejectsStreamBeforeUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstreamCalled := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()
	client, err := grok.NewClient(upstream.URL, "key", "grok-4")
	if err != nil {
		t.Fatal(err)
	}
	api := &agentAPI{grok: client}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/agent/v1/grok/responses", strings.NewReader(`{"model":"grok-4","input":"hi","stream":true}`))
	context.Request.Header.Set("Content-Type", "application/json")
	api.grokResponses(context)
	if recorder.Code != http.StatusNotImplemented {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if upstreamCalled {
		t.Fatal("stream rejection must happen before upstream")
	}
}

func TestNewAgentGrokClientReadsModelAllowlist(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_GROK_BASE_URL", "https://api.x.ai/v1")
	t.Setenv("OLLAMA_AGENT_GROK_MODEL", "grok-4")
	t.Setenv("OLLAMA_AGENT_GROK_MODELS", "grok-4,grok-4.1")
	client, err := newAgentGrokClient()
	if err != nil {
		t.Fatal(err)
	}
	if len(client.AllowedModels) != 2 || client.AllowedModels[1] != "grok-4.1" {
		t.Fatalf("allowlist=%v", client.AllowedModels)
	}
}
