package agent

import (
	"errors"
	"testing"
)

func TestDeploymentApprovalPersistsAcrossRestartAndConsumesOnce(t *testing.T) {
	root := t.TempDir()
	store, err := NewDeploymentApprovalStore(root)
	if err != nil {
		t.Fatal(err)
	}
	requested, err := store.Request("org_a", "bld_a", "self", "staging", "operator_a")
	if err != nil {
		t.Fatal(err)
	}
	if requested.Status != DeploymentApprovalPending || requested.Nonce == "" {
		t.Fatalf("requested approval = %+v", requested)
	}
	if _, err := store.Decide(requested.ID, "org_a", "bld_a", "self", "operator_a", "self approved", requested.Nonce, true); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Decide(requested.ID, "org_a", "bld_a", "self", "operator_a", "replay", requested.Nonce, true); !errors.Is(err, ErrDeploymentApprovalConflict) {
		t.Fatalf("repeated decision error = %v", err)
	}

	reloaded, err := NewDeploymentApprovalStore(root)
	if err != nil {
		t.Fatal(err)
	}
	consumed, err := reloaded.Consume(requested.ID, "org_a", "bld_a", "self", "staging", requested.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	if consumed.Status != DeploymentApprovalConsumed || consumed.DecidedBy != "operator_a" {
		t.Fatalf("consumed approval = %+v", consumed)
	}
	if _, err := reloaded.Consume(requested.ID, "org_a", "bld_a", "self", "staging", requested.Nonce); !errors.Is(err, ErrDeploymentApprovalConflict) {
		t.Fatalf("replay consume error = %v", err)
	}
}

func TestDeploymentApprovalRejectsWrongTenantAndNonce(t *testing.T) {
	store, err := NewDeploymentApprovalStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	approval, err := store.Request("org_a", "bld_a", "self", "production", "operator_a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Decide(approval.ID, "org_b", "bld_a", "self", "admin_b", "approve", approval.Nonce, true); !errors.Is(err, ErrDeploymentApprovalOrganization) {
		t.Fatalf("wrong tenant decision error = %v", err)
	}
	if _, err := store.Decide(approval.ID, "org_a", "bld_a", "self", "admin_a", "approve", "wrong", true); !errors.Is(err, ErrDeploymentApprovalNonce) {
		t.Fatalf("wrong nonce decision error = %v", err)
	}
}
