package agent

import (
	"errors"
	"testing"
)

func TestCompanyTelAgentIdempotencyReplaysWithoutDuplicateMutation(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Idempotent Co"})
	if err != nil {
		t.Fatal(err)
	}
	request := TelAgentRequest{Message: "criar tarefa", Operation: "backlog.create", Title: "Uma tarefa", Priority: 10, IdempotencyKey: "tel-agent-1"}
	first, firstResult, err := store.ExecuteTelAgent(company.ID, "org-a", "user-a", request)
	if err != nil {
		t.Fatal(err)
	}
	second, secondResult, err := store.ExecuteTelAgent(company.ID, "org-a", "user-a", request)
	if err != nil {
		t.Fatal(err)
	}
	if firstResult.Exchange.ID != secondResult.Exchange.ID || firstResult.Exchange.CreatedResourceID != secondResult.Exchange.CreatedResourceID {
		t.Fatalf("replay created a new exchange: first=%+v second=%+v", firstResult.Exchange, secondResult.Exchange)
	}
	if len(first.Backlog) != 1 || len(second.Backlog) != 1 || len(second.TelAgentHistory) != 1 {
		t.Fatalf("replay duplicated mutation: first=%+v second=%+v", first, second)
	}
}

func TestCompanyTelAgentIdempotencyRejectsDifferentPayload(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Conflict Co"})
	if err != nil {
		t.Fatal(err)
	}
	key := "tel-agent-conflict"
	_, _, err = store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: "criar tarefa", Operation: "backlog.create", Title: "Primeira", IdempotencyKey: key})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.ExecuteTelAgent(company.ID, "org-a", "user-a", TelAgentRequest{Message: "criar tarefa", Operation: "backlog.create", Title: "Outra", IdempotencyKey: key})
	if !errors.Is(err, ErrCompanyIdempotencyConflict) {
		t.Fatalf("conflict err=%v", err)
	}
}
