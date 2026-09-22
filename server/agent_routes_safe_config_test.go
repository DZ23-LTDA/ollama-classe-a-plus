package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSafeConfigRedactsValuesAndReportsConfiguredStates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("OLLAMA_AGENT_MODEL", "qwen-test")
	t.Setenv("OLLAMA_AGENT_DATABASE_URL", "postgres://secret-user:secret-pass@db.internal/agent")
	t.Setenv("OLLAMA_AGENT_REDIS_URL", "redis://secret:secret@redis.internal/0")
	t.Setenv("OLLAMA_AGENT_CONNECTORS", "/private/connectors.json")
	t.Setenv("OLLAMA_AGENT_MEDIA_BASE_URL", "https://media.internal")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	(&agentAPI{authRequired: true}).safeConfig(context)

	if recorder.Code != 200 {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["store"] != "postgres" || body["queue"] != "redis" || body["planner_model_configured"] != true {
		t.Fatalf("unexpected safe config: %#v", body)
	}
	for _, secret := range []string{"secret-user", "secret-pass", "private/connectors.json", "media.internal", "qwen-test"} {
		if containsString(recorder.Body.String(), secret) {
			t.Fatalf("safe config leaked %q: %s", secret, recorder.Body.String())
		}
	}
}

func containsString(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
