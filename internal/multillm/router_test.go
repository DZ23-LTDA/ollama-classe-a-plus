package multillm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRouteSelectsHealthyModelWithinConstraints(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.json")
	data := `{"providers":[{"name":"cheap","type":"openai-compatible","base_url":"https://cheap.example","api_key_env":"CHEAP_KEY","priority":1,"models":[{"id":"small","capabilities":["coding"],"cost_per_1k_input_cents":1,"cost_per_1k_output_cents":1,"quality_score":5}]},{"name":"quality","type":"openai-compatible","base_url":"https://quality.example","api_key_env":"QUALITY_KEY","priority":2,"models":[{"id":"large","capabilities":["coding"],"cost_per_1k_input_cents":5,"cost_per_1k_output_cents":5,"quality_score":10}]}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHEAP_KEY", "cheap")
	t.Setenv("QUALITY_KEY", "quality")
	registry, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := registry.Route(RouteRequest{RequiredCapabilities: []string{"coding"}, MaxCostCents: 10, Health: map[string]ProviderHealth{"cheap": {Healthy: true, LatencyMS: 300}, "quality": {Healthy: true, LatencyMS: 50}}})
	if err != nil || decision.Model.ID != "quality/large" {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
}

func TestRouteRejectsUnhealthyOrTooSlowProviders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.json")
	data := `{"providers":[{"name":"slow","type":"openai-compatible","base_url":"https://slow.example","models":[{"id":"model","capabilities":["reasoning"]}]}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	registry, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Route(RouteRequest{RequiredCapabilities: []string{"reasoning"}, MaxLatencyMS: 100, Health: map[string]ProviderHealth{"slow": {Healthy: true, LatencyMS: 101}}}); err != ErrNoRoute {
		t.Fatalf("expected no route for latency, got %v", err)
	}
}
