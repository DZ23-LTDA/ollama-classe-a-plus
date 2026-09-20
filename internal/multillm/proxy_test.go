package multillm

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProxyRoutesConfiguredModelAndRedactsClientAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var gotModel, gotAuthorization string
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		gotModel, _ = body["model"].(string)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	defer upstream.Close()

	t.Setenv("REMOTE_KEY", "provider-secret")
	r := &Registry{
		providers: map[string]Provider{"remote": {Name: "remote", Type: "openai-compatible", BaseURL: upstream.URL + "/v1", APIKeyEnv: "REMOTE_KEY", AllowPrivate: true}},
		models:    map[string]Model{"remote/model": {ID: "remote/model", UpstreamID: "upstream-model", Provider: "remote", Available: true}},
	}
	g := NewGateway(r, upstream.Client())
	router := gin.New()
	router.Use(g.Middleware())
	router.POST("/v1/chat/completions", func(c *gin.Context) { t.Fatal("request was not proxied") })

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"remote/model","messages":[]}`))
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Authorization", "Bearer client-secret")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || gotModel != "upstream-model" {
		t.Fatalf("status=%d upstream model=%q body=%s", rec.Code, gotModel, rec.Body.String())
	}
	if gotAuthorization != "Bearer provider-secret" {
		t.Fatalf("upstream authorization = %q", gotAuthorization)
	}
}

func TestCLIProviderRequiresExplicitExecutionAndReturnsCompletion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DZ23_CLI_HELPER", "1")
	r := &Registry{
		providers: map[string]Provider{"codex": {Name: "codex", Type: "cli", Executable: os.Args[0], Args: []string{"-test.run=TestCLIHelperProcess"}, AllowExecution: true}},
		models:    map[string]Model{"codex/cli": {ID: "codex/cli", UpstreamID: "cli", Provider: "codex", Available: true}},
	}
	router := gin.New()
	router.Use(NewGateway(r, nil).Middleware())
	router.POST("/v1/chat/completions", func(c *gin.Context) { t.Fatal("request was not executed") })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"codex/cli","messages":[{"role":"user","content":"hello"}]}`))
	request.RemoteAddr = "127.0.0.1:12345"
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "assistant: user: hello") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCLIProviderAcceptsNativeGeneratePrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DZ23_CLI_HELPER", "1")
	r := &Registry{
		providers: map[string]Provider{"cli": {Name: "cli", Type: ProviderTypeCLI, Executable: os.Args[0], Args: []string{"-test.run=TestCLIHelperProcess"}, AllowExecution: true}},
		models:    map[string]Model{"cli/default": {ID: "cli/default", UpstreamID: "default", Provider: "cli", Available: true}},
	}
	router := gin.New()
	router.Use(NewGateway(r, nil).Middleware())
	router.POST("/api/generate", func(c *gin.Context) { t.Fatal("request was not executed") })
	request := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(`{"model":"cli/default","stream":false,"prompt":"hello generate"}`))
	request.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"response":"assistant: hello generate"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAutoRoutingSkipsProviderWithoutRequestedPath(t *testing.T) {
	r := &Registry{
		providers: map[string]Provider{
			"chat-only": {Name: "chat-only", Type: ProviderTypeOpenAICompatible, Paths: []string{"/v1/chat/completions"}},
			"responses": {Name: "responses", Type: ProviderTypeOpenAICompatible, Paths: []string{"/v1/responses"}},
		},
		models: map[string]Model{
			"chat-only/model": {ID: "chat-only/model", Provider: "chat-only", Available: true, Priority: 100, Capabilities: []string{"coding"}},
			"responses/model": {ID: "responses/model", Provider: "responses", Available: true, Priority: 10, Capabilities: []string{"coding"}},
		},
	}
	model, ok := r.Resolve("auto/coding", Policy{Path: "/v1/responses"})
	if !ok || model.ID != "responses/model" {
		t.Fatalf("resolved %#v, ok=%v", model, ok)
	}
}

func TestBoundedReaderReturnsExplicitLimitError(t *testing.T) {
	reader := &boundedReader{reader: strings.NewReader("abcd"), remaining: 3}
	_, err := io.ReadAll(reader)
	if err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("error = %v", err)
	}
}

func TestOversizedSSEEmitsClientVisibleErrorEvent(t *testing.T) {
	var output strings.Builder
	err := copySSEBounded(&output, strings.NewReader("1234"), 3)
	if !errors.Is(err, errResponseLimit) {
		t.Fatalf("error = %v", err)
	}
	if output.String() != "123\nevent: error\ndata: {\"error\":{\"message\":\"provider response exceeded limit\",\"type\":\"response_too_large\"}}\n\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRemoteCallerCannotSpendProviderKeyWithoutGatewayAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := &Registry{
		providers: map[string]Provider{"remote": {Name: "remote", Type: "openai-compatible", BaseURL: "https://example.com/v1"}},
		models:    map[string]Model{"remote/model": {ID: "remote/model", UpstreamID: "model", Provider: "remote", Available: true}},
	}
	router := gin.New()
	router.Use(NewGateway(r, nil).Middleware())
	router.POST("/v1/chat/completions", func(c *gin.Context) { t.Fatal("request bypassed authorization") })
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"remote/model","messages":[]}`))
	request.RemoteAddr = "192.0.2.10:12345"
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCLIHelperProcess(t *testing.T) {
	if os.Getenv("DZ23_CLI_HELPER") != "1" {
		return
	}
	b, _ := io.ReadAll(os.Stdin)
	_, _ = os.Stdout.WriteString("assistant: " + strings.TrimSpace(string(b)))
	os.Exit(0)
}

func TestProxyLeavesLocalModelsForOllama(t *testing.T) {
	r := NewGateway(&Registry{providers: map[string]Provider{}, models: map[string]Model{}}, http.DefaultClient)
	router := gin.New()
	router.Use(r.Middleware())
	router.POST("/v1/chat/completions", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"qwen:latest"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestNativeChatRoutesRemoteModelAndReturnsOllamaShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"remote answer"}}]}`))
	}))
	defer upstream.Close()
	r := &Registry{
		providers: map[string]Provider{"remote": {Name: "remote", Type: "openai-compatible", BaseURL: upstream.URL + "/v1", AllowPrivate: true}},
		models:    map[string]Model{"remote/model": {ID: "remote/model", UpstreamID: "upstream-model", Provider: "remote", Available: true}},
	}
	router := gin.New()
	router.Use(NewGateway(r, upstream.Client()).Middleware())
	router.POST("/api/chat", func(c *gin.Context) { t.Fatal("request was not proxied") })
	request := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"model":"remote/model","stream":false,"messages":[{"role":"user","content":"hello"}]}`))
	request.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"content":"remote answer"`) || !strings.Contains(recorder.Body.String(), `"done":true`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
