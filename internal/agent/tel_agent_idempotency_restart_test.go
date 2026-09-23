package agent

import "testing"

func TestCompanyTelAgentIdempotencySurvivesRestart(t *testing.T) {
	root := t.TempDir()
	store, err := NewCompanyStore(root)
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Restart Co"})
	if err != nil {
		t.Fatal(err)
	}
	request := TelAgentRequest{Message: "criar tarefa", Operation: "backlog.create", Title: "Persistente", IdempotencyKey: "restart-key"}
	first, firstResult, err := store.ExecuteTelAgent(company.ID, "org-a", "user-a", request)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewCompanyStore(root)
	if err != nil {
		t.Fatal(err)
	}
	second, secondResult, err := reloaded.ExecuteTelAgent(company.ID, "org-a", "user-a", request)
	if err != nil {
		t.Fatal(err)
	}
	if firstResult.Exchange.ID != secondResult.Exchange.ID || firstResult.Exchange.CreatedResourceID != secondResult.Exchange.CreatedResourceID {
		t.Fatalf("restart replay changed result: first=%+v second=%+v", firstResult.Exchange, secondResult.Exchange)
	}
	if len(first.Backlog) != 1 || len(second.Backlog) != 1 || len(second.TelAgentHistory) != 1 {
		t.Fatalf("restart replay duplicated mutation: first=%+v second=%+v", first, second)
	}
}
