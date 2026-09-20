package multillm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRegistryPublishesConfiguredModelsWithoutSecrets(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-secret")
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.json")
	config := `{
	  "providers": [{
	    "name": "openai",
	    "type": "openai-compatible",
	    "base_url": "https://api.openai.com/v1",
	    "api_key_env": "OPENAI_API_KEY",
	    "models": [{"id":"gpt-test","capabilities":["chat","tools"]}]
	  }]
	}`
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	r, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	models := r.Models()
	if len(models) != 1 || models[0].ID != "openai/gpt-test" {
		t.Fatalf("models = %#v", models)
	}
	if models[0].Provider != "openai" || !models[0].Available {
		t.Fatalf("model metadata = %#v", models[0])
	}
	if got := r.SafeSnapshot(); contains(got, "test-secret") {
		t.Fatalf("safe snapshot leaked API key: %s", got)
	}
}

func TestLoadRegistryRejectsUnsafeProviderURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.json")
	if err := os.WriteFile(path, []byte(`{"providers":[{"name":"bad","type":"openai-compatible","base_url":"http://example.com/v1","models":[{"id":"x"}]}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected insecure provider URL to be rejected")
	}
}

func TestResolveNeverFallsBackForLocalOnly(t *testing.T) {
	r := &Registry{models: map[string]Model{
		"openai/gpt-test": {ID: "openai/gpt-test", Provider: "openai", Available: true},
	}}
	if _, ok := r.Resolve("local/private", Policy{LocalOnly: true}); ok {
		t.Fatal("local/private must not resolve to a remote model")
	}
}

func contains(s, needle string) bool {
	for i := 0; i+len(needle) <= len(s); i++ {
		if s[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
