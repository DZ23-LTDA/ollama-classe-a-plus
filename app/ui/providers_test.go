//go:build windows || darwin

package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeProviderConfig(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "providers.json")
	cfg := `{"providers":[
	  {"name":"openai","type":"openai-compatible","base_url":"https://api.openai.com/v1","api_key_env":"DZ23_TEST_OPENAI_KEY","models":[{"id":"a"},{"id":"b"}]},
	  {"name":"local","type":"openai-compatible","base_url":"http://127.0.0.1:1234/v1","models":[{"id":"c"}]}
	]}`
	if err := os.WriteFile(path, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OLLAMA_DZ23_CONFIG", path)
	t.Setenv("DZ23_TEST_OPENAI_KEY", "")
	t.Setenv("DZ23_TEST_OPENAI_KEY_FILE", "")
}

func providerRequest(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "token", Value: "t"})
	rr := httptest.NewRecorder()
	(&Server{Token: "t"}).Handler().ServeHTTP(rr, req)
	return rr
}

func TestListProvidersReportsStatusWithoutSecrets(t *testing.T) {
	writeProviderConfig(t)
	t.Setenv("DZ23_TEST_OPENAI_KEY", "sk-should-never-be-returned")

	rr := providerRequest(t, http.MethodGet, "/api/v1/providers", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "sk-should-never-be-returned") {
		t.Fatal("provider listing leaked a credential")
	}
	var got struct {
		Providers []ProviderStatus `json:"providers"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Providers) != 2 {
		t.Fatalf("providers = %+v", got.Providers)
	}
	openai, local := got.Providers[0], got.Providers[1]
	if !openai.NeedsKey || !openai.Configured || openai.Models != 2 {
		t.Fatalf("openai = %+v", openai)
	}
	if local.NeedsKey || !local.Configured {
		t.Fatalf("keyless provider = %+v", local)
	}
}

func TestProviderKeyRejectsUnknownProviderAndBadKeys(t *testing.T) {
	writeProviderConfig(t)

	if rr := providerRequest(t, http.MethodPut, "/api/v1/providers/nope/key", `{"key":"x"}`); rr.Code != http.StatusNotFound {
		t.Fatalf("unknown provider status = %d", rr.Code)
	}
	if rr := providerRequest(t, http.MethodPut, "/api/v1/providers/local/key", `{"key":"x"}`); rr.Code != http.StatusNotFound {
		t.Fatalf("keyless provider status = %d", rr.Code)
	}
	if rr := providerRequest(t, http.MethodPut, "/api/v1/providers/openai/key", `{"key":"two\nlines"}`); rr.Code != http.StatusBadRequest {
		t.Fatalf("bad key status = %d", rr.Code)
	}
}

func TestProviderEndpointsRequireUIToken(t *testing.T) {
	writeProviderConfig(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/providers", nil)
	rr := httptest.NewRecorder()
	(&Server{Token: "t"}).Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status without token = %d, want 403", rr.Code)
	}
}
