package agent

import (
	"errors"
	"testing"
)

func TestTelAgentSessionHandoffJourneyAndPersistence(t *testing.T) {
	root := t.TempDir()
	store, err := NewCompanyStore(root)
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "Suporte Co"})
	if err != nil {
		t.Fatal(err)
	}

	// Start -> ACTIVE_AI
	_, session, err := store.StartTelAgentSession(company.ID, "org-a", "cliente-1", "whatsapp")
	if err != nil || session.State != TelSessionActiveAI || session.ActorID != "cliente-1" {
		t.Fatalf("start: session=%+v err=%v", session, err)
	}
	sid := session.ID

	// actor obrigatorio
	if _, _, err := store.StartTelAgentSession(company.ID, "org-a", "  ", "x"); !errors.Is(err, ErrTelSessionActorRequired) {
		t.Fatalf("actor required err=%v", err)
	}

	// Request handoff -> WAITING_HUMAN
	if _, s, err := store.RequestTelAgentHandoff(company.ID, "org-a", sid); err != nil || s.State != TelSessionWaitingHuman {
		t.Fatalf("handoff: %+v err=%v", s, err)
	}

	// Handoff sem operador -> erro
	if _, _, err := store.TelAgentHumanTakeOver(company.ID, "org-a", sid, "  "); !errors.Is(err, ErrTelSessionHumanMissing) {
		t.Fatalf("handoff sem operador err=%v", err)
	}

	// Human takeover -> HUMAN_ACTIVE
	if _, s, err := store.TelAgentHumanTakeOver(company.ID, "org-a", sid, "operador-joao"); err != nil || s.State != TelSessionHumanActive || s.HumanOwner != "operador-joao" {
		t.Fatalf("takeover: %+v err=%v", s, err)
	}

	// Dupla tomada -> erro
	if _, _, err := store.TelAgentHumanTakeOver(company.ID, "org-a", sid, "operador-maria"); !errors.Is(err, ErrTelSessionAlreadyHuman) {
		t.Fatalf("dupla tomada err=%v", err)
	}

	// Request handoff invalido enquanto HUMAN_ACTIVE
	if _, _, err := store.RequestTelAgentHandoff(company.ID, "org-a", sid); !errors.Is(err, ErrTelSessionInvalidState) {
		t.Fatalf("handoff invalido err=%v", err)
	}

	// Human return -> RESUMED_AI
	if _, s, err := store.TelAgentHumanReturn(company.ID, "org-a", sid); err != nil || s.State != TelSessionResumedAI || s.HumanOwner != "" {
		t.Fatalf("return: %+v err=%v", s, err)
	}

	// Persistencia / restart: recarrega o store e confere o estado.
	reloaded, err := NewCompanyStore(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := reloaded.TelAgentSessionForOrganization(company.ID, "org-a", sid)
	if err != nil || got.State != TelSessionResumedAI {
		t.Fatalf("persistencia: got=%+v err=%v", got, err)
	}

	// Close -> CLOSED, e operacoes seguintes falham.
	if _, s, err := reloaded.CloseTelAgentSession(company.ID, "org-a", sid); err != nil || s.State != TelSessionClosed {
		t.Fatalf("close: %+v err=%v", s, err)
	}
	if _, _, err := reloaded.RequestTelAgentHandoff(company.ID, "org-a", sid); !errors.Is(err, ErrTelSessionClosed) {
		t.Fatalf("op em sessao fechada err=%v", err)
	}
}

func TestTelAgentSessionRejectsCrossTenant(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org-a", Name: "A Co"})
	if err != nil {
		t.Fatal(err)
	}
	_, session, err := store.StartTelAgentSession(company.ID, "org-a", "cli", "chat")
	if err != nil {
		t.Fatal(err)
	}
	sid := session.ID

	// ORG_B nao pode ler nem transicionar a sessao de ORG_A.
	if _, err := store.TelAgentSessionForOrganization(company.ID, "org-b", sid); !errors.Is(err, ErrTelAgentOrganizationMismatch) {
		t.Fatalf("read cross-tenant err=%v", err)
	}
	if _, _, err := store.RequestTelAgentHandoff(company.ID, "org-b", sid); !errors.Is(err, ErrTelAgentOrganizationMismatch) {
		t.Fatalf("handoff cross-tenant err=%v", err)
	}
	if _, _, err := store.TelAgentHumanTakeOver(company.ID, "org-b", sid, "intruso"); !errors.Is(err, ErrTelAgentOrganizationMismatch) {
		t.Fatalf("takeover cross-tenant err=%v", err)
	}
	if _, _, err := store.CloseTelAgentSession(company.ID, "org-b", sid); !errors.Is(err, ErrTelAgentOrganizationMismatch) {
		t.Fatalf("close cross-tenant err=%v", err)
	}

	// A sessao de ORG_A permanece intacta (ACTIVE_AI).
	own, err := store.TelAgentSessionForOrganization(company.ID, "org-a", sid)
	if err != nil || own.State != TelSessionActiveAI {
		t.Fatalf("sessao do dono alterada: %+v err=%v", own, err)
	}
}
