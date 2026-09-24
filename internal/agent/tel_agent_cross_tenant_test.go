package agent

import (
	"errors"
	"testing"
)

// ORG_A jamais pode ler o historico tel-agent de ORG_B, e a negativa nao pode
// vazar dado nem inferir existencia. Complementa o negative test de Execute em
// tel_agent_test.go, cobrindo o caminho de LEITURA (history).
func TestCompanyTelAgentHistoryRejectsCrossTenantWithoutLeak(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Org A Co"})
	if err != nil {
		t.Fatal(err)
	}
	// Gera historico real no tenant dono.
	if _, _, err := store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: "estado", Operation: "report.read"}); err != nil {
		t.Fatal(err)
	}

	// ORG_B tentando ler o historico de ORG_A: deve falhar com mismatch e
	// retornar historico nil (nenhum dado vazado).
	history, err := store.TelAgentHistoryForOrganization(company.ID, "org-b", 10)
	if !errors.Is(err, ErrTelAgentOrganizationMismatch) {
		t.Fatalf("cross-tenant history esperava mismatch, got err=%v", err)
	}
	if history != nil {
		t.Fatalf("cross-tenant history vazou dados: %+v", history)
	}

	// Controle positivo: o dono le normalmente.
	own, err := store.TelAgentHistoryForOrganization(company.ID, "org-a", 10)
	if err != nil || len(own) != 1 {
		t.Fatalf("dono deveria ler o proprio historico: err=%v history=%+v", err, own)
	}
}

// Uma tentativa cross-tenant de Execute nao pode causar side effect (nenhuma
// mutacao no company do tenant dono).
func TestCompanyTelAgentCrossTenantExecuteHasNoSideEffect(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "NoSideEffect Co"})
	if err != nil {
		t.Fatal(err)
	}

	before, err := store.Get(company.ID)
	if err != nil {
		t.Fatal(err)
	}

	// ORG_B tenta criar backlog no company de ORG_A.
	if _, _, err := store.ExecuteTelAgent(company.ID, "org-b", "user-b", TelAgentRequest{Message: "criar tarefa", Operation: "backlog.create", Title: "intruso", Description: "nao deveria existir", Priority: 5}); !errors.Is(err, ErrTelAgentOrganizationMismatch) {
		t.Fatalf("esperava mismatch no cross-tenant execute, got err=%v", err)
	}

	after, err := store.Get(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Backlog) != len(before.Backlog) || len(after.TelAgentHistory) != len(before.TelAgentHistory) || len(after.Approvals) != len(before.Approvals) {
		t.Fatalf("cross-tenant execute causou side effect: before backlog=%d hist=%d appr=%d / after backlog=%d hist=%d appr=%d",
			len(before.Backlog), len(before.TelAgentHistory), len(before.Approvals),
			len(after.Backlog), len(after.TelAgentHistory), len(after.Approvals))
	}
}
