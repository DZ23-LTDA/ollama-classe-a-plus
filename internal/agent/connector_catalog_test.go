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

func TestConnectorCatalogIDsAreUniqueAndSetupIsExplicit(t *testing.T) {
	seen := map[string]bool{}
	for _, entry := range ConnectorCatalog() {
		if seen[entry.ID] {
			t.Fatalf("duplicate connector id %q", entry.ID)
		}
		seen[entry.ID] = true
		if entry.Name == "" || entry.Description == "" || entry.Auth == "" {
			t.Fatalf("connector %q is missing name, description or auth", entry.ID)
		}
	}
	for _, id := range []string{"pinterest", "youtube", "linkedin", "telegram", "gmail", "mercado-pago", "resend", "twilio", "firebase", "sentry", "aws"} {
		if !seen[id] {
			t.Errorf("catalog is missing %q", id)
		}
	}
}
