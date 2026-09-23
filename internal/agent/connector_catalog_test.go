package agent

import "testing"

func TestConnectorCatalogIncludesOperationalIntegrations(t *testing.T) {
	entries := ConnectorCatalog()
	if len(entries) < 20 {
		t.Fatalf("expected broad connector catalog, got %d entries", len(entries))
	}
	wanted := map[string]bool{
		"composio":         false,
		"google-workspace": false,
		"github":           false,
		"woovi-openpix":    false,
		"fiscal-invoicing": false,
		"shopify":          false,
		"tiktok-business":  false,
		"vercel":           false,
	}
	for _, entry := range entries {
		if _, ok := wanted[entry.ID]; ok {
			wanted[entry.ID] = true
		}
		if entry.Status == "available" && entry.Auth == "" {
			t.Fatalf("available catalog entry %q has no auth contract", entry.ID)
		}
	}
	for id, found := range wanted {
		if !found {
			t.Errorf("catalog is missing %q", id)
		}
	}
}

func TestConnectorCatalogDoesNotClaimFiscalIssuerByDefault(t *testing.T) {
	for _, entry := range ConnectorCatalog() {
		if entry.ID == "fiscal-invoicing" && entry.Status != "provider_selection_required" {
			t.Fatalf("fiscal catalog entry must require provider selection, got %q", entry.Status)
		}
	}
}
