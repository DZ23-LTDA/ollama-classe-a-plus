//go:build windows || darwin

package ui

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ollama/ollama/app/secrets"
	"github.com/ollama/ollama/internal/multillm"
)

func TestConnectConnectorRegistersWithTokenEnv(t *testing.T) {
	var registered map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/agent/v1/connectors" && r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &registered)
			w.Write([]byte(`{}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer upstream.Close()
	t.Setenv("OLLAMA_HOST", upstream.URL)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	t.Setenv("OLLAMA_CONNECTOR_RESEND_TOKEN", "")
	t.Setenv("OLLAMA_CONNECTOR_RESEND_TOKEN_FILE", "")
	t.Cleanup(func() { _ = removeTestConnectorEnv("OLLAMA_CONNECTOR_RESEND_TOKEN") })

	rr := providerRequest(t, http.MethodPut, "/api/v1/connectors/resend/key", `{"key":"re_test_123"}`)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	if registered["id"] != "resend" || registered["base_url"] != "https://api.resend.com" || registered["token_env"] != "OLLAMA_CONNECTOR_RESEND_TOKEN" {
		t.Fatalf("registered = %v", registered)
	}
	if _, leaked := registered["key"]; leaked {
		t.Fatal("the key must never be sent to the agent API")
	}
}

func TestConnectConnectorErrors(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"bearer token is required"}`, http.StatusUnauthorized)
	}))
	defer upstream.Close()
	t.Setenv("OLLAMA_HOST", upstream.URL)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	t.Setenv("OLLAMA_CONNECTOR_GITHUB_TOKEN", "")
	t.Setenv("OLLAMA_CONNECTOR_GITHUB_TOKEN_FILE", "")
	t.Cleanup(func() { _ = removeTestConnectorEnv("OLLAMA_CONNECTOR_GITHUB_TOKEN") })

	if rr := providerRequest(t, http.MethodPut, "/api/v1/connectors/gmail/key", `{"key":"x"}`); rr.Code != http.StatusNotFound {
		t.Fatalf("oauth-only connector status = %d", rr.Code)
	}
	if rr := providerRequest(t, http.MethodPut, "/api/v1/connectors/coolify/key", `{"key":"x","base_url":"http://insecure"}`); rr.Code != http.StatusBadRequest {
		t.Fatalf("self-hosted without https status = %d", rr.Code)
	}
	if rr := providerRequest(t, http.MethodPut, "/api/v1/connectors/github/key", `{"key":""}`); rr.Code != http.StatusBadRequest {
		t.Fatalf("empty key status = %d", rr.Code)
	}
	if rr := providerRequest(t, http.MethodPut, "/api/v1/connectors/github/key", `{"key":"ghp_x"}`); rr.Code != http.StatusForbidden {
		t.Fatalf("login-required status = %d, body %s", rr.Code, rr.Body.String())
	}
}

// removeTestConnectorEnv clears the saved key and the persisted user
// variable that secrets.Save writes, so tests leave no trace.
func removeTestConnectorEnv(envName string) error {
	return secrets.Remove(multillm.DefaultCredentialDir(), envName)
}

func TestConnectConnectorSendsHeaderAndSelfHostedURL(t *testing.T) {
	var registered map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &registered)
		w.Write([]byte(`{}`))
	}))
	defer upstream.Close()
	t.Setenv("OLLAMA_HOST", upstream.URL)
	t.Setenv("LOCALAPPDATA", t.TempDir())
	for _, env := range []string{"OLLAMA_CONNECTOR_APPWRITE_TOKEN", "OLLAMA_CONNECTOR_COOLIFY_TOKEN"} {
		t.Setenv(env, "")
		t.Setenv(env+"_FILE", "")
		t.Cleanup(func() { _ = removeTestConnectorEnv(env) })
	}

	if rr := providerRequest(t, http.MethodPut, "/api/v1/connectors/appwrite/key", `{"key":"k"}`); rr.Code != http.StatusNoContent {
		t.Fatalf("appwrite status %d: %s", rr.Code, rr.Body.String())
	}
	if registered["auth_header"] != "X-Appwrite-Key" || registered["base_url"] != "https://cloud.appwrite.io/v1" {
		t.Fatalf("appwrite registered = %v", registered)
	}

	if rr := providerRequest(t, http.MethodPut, "/api/v1/connectors/coolify/key", `{"key":"k","base_url":"https://coolify.example.com/api/v1/"}`); rr.Code != http.StatusNoContent {
		t.Fatalf("coolify status %d: %s", rr.Code, rr.Body.String())
	}
	if registered["base_url"] != "https://coolify.example.com/api/v1" {
		t.Fatalf("coolify registered = %v", registered)
	}
}
