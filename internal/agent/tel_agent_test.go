package agent

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCompanyTelAgentExecutesAllowlistedOperationsAndPersistsHistory(t *testing.T) {
	root := t.TempDir()
	store, err := NewCompanyStore(root)
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Local Co", Budget: CompanyBudget{MonthlyLimitCents: 10000, SpentCents: 1000}})
	if err != nil {
		t.Fatal(err)
	}

	updated, result, err := store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: "criar tarefa para revisar onboarding", Operation: "backlog.create", Title: "Revisar onboarding", Description: "Testar o fluxo local", Priority: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result.Exchange.Channel != "tel-agent.text" || result.Exchange.Status != "completed" || result.Exchange.CreatedResourceID == "" {
		t.Fatalf("unexpected backlog exchange=%+v", result.Exchange)
	}
	if len(updated.Backlog) != 1 || updated.Backlog[0].Source != "tel-agent" || len(updated.TelAgentHistory) != 1 {
		t.Fatalf("backlog/history not persisted: %+v", updated)
	}

	updated, result, err = store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: "preparar campanha de lançamento", Operation: "campaign.draft", Title: "Lançamento", Description: "Rascunho local", DailyBudgetCents: 500})
	if err != nil {
		t.Fatal(err)
	}
	if result.Exchange.Status != "approval_pending" || !result.Exchange.ApprovalRequired || len(updated.Campaigns) != 1 || len(updated.Approvals) != 1 {
		t.Fatalf("campaign did not require approval: exchange=%+v company=%+v", result.Exchange, updated)
	}

	_, result, err = store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: "qual o estado atual?", Operation: "report.read"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Report == nil || result.Report.OpenBacklog != 1 || !strings.Contains(result.Exchange.Reply, "Local Co") {
		t.Fatalf("report exchange=%+v", result)
	}

	reloaded, err := NewCompanyStore(root)
	if err != nil {
		t.Fatal(err)
	}
	history, err := reloaded.TelAgentHistoryForOrganization(company.ID, "org-a", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 3 || history[0].ActorID != "user-a" {
		t.Fatalf("reloaded history=%+v", history)
	}
	if _, err := os.Stat(root + "/" + company.ID + ".json"); err != nil {
		t.Fatal(err)
	}
}

func TestCompanyTelAgentRejectsCrossTenantAndUnsupportedInput(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Scoped Co"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ExecuteTelAgent(company.ID, "org-b", "user-b", TelAgentRequest{Message: "read", Operation: "report.read"}); !errors.Is(err, ErrTelAgentOrganizationMismatch) {
		t.Fatalf("cross-tenant err=%v", err)
	}
	if _, _, err := store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: "read", Operation: "shell.exec"}); !errors.Is(err, ErrTelAgentUnsupportedOperation) {
		t.Fatalf("unsupported operation err=%v", err)
	}
	if _, _, err := store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: strings.Repeat("x", maxTelAgentMessageBytes+1), Operation: "report.read"}); !errors.Is(err, ErrTelAgentMessageTooLong) {
		t.Fatalf("oversized message err=%v", err)
	}
}

func TestCompanyTelAgentHistoryIsBounded(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Bounded Co"})
	if err != nil {
		t.Fatal(err)
	}
	for range maxTelAgentHistory + 10 {
		if _, _, err := store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: "state", Operation: "report.read"}); err != nil {
			t.Fatal(err)
		}
	}
	updated, err := store.Get(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.TelAgentHistory) != maxTelAgentHistory {
		t.Fatalf("history len=%d, want %d", len(updated.TelAgentHistory), maxTelAgentHistory)
	}
}
