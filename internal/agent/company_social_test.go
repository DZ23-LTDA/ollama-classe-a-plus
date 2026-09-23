package agent

import (
	"errors"
	"testing"
)

func TestCompanySocialApprovalAndSandboxPublish(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{Name: "Social test", OrganizationID: "org_social"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddSocialAccount(company.ID, CompanySocialAccount{Provider: "unknown", Name: "bad"}); !errors.Is(err, ErrCompanySocialProviderUnsupported) {
		t.Fatalf("unsupported provider err=%v", err)
	}
	company, err = store.AddSocialAccount(company.ID, CompanySocialAccount{Provider: "x", Name: "brand"})
	if err != nil || len(company.SocialAccounts) != 1 || company.SocialAccounts[0].Status != "pending_oauth" {
		t.Fatalf("account=%+v err=%v", company.SocialAccounts, err)
	}
	company, err = store.CreateSocialDraft(company.ID, CompanySocialDraft{Provider: "x", Title: "draft", Body: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	draftID := company.SocialDrafts[0].ID
	if _, err := store.PublishSocialDraft(company.ID, draftID); !errors.Is(err, ErrCompanyApprovalRequiredForExternal) {
		t.Fatalf("publish before approval err=%v", err)
	}
	if _, err := store.ApproveSocialDraft(company.ID, draftID); err != nil {
		t.Fatal(err)
	}
	company, err = store.PublishSocialDraft(company.ID, draftID)
	if err != nil || company.SocialDrafts[0].Status != "published_sandbox" {
		t.Fatalf("published=%+v err=%v", company.SocialDrafts, err)
	}
	company, err = store.RecordSocialMetric(company.ID, CompanySocialMetric{DraftID: draftID, Provider: "x", Impressions: 100, Clicks: 10, Conversions: 2})
	if err != nil {
		t.Fatal(err)
	}
	report, err := store.SocialReport(company.ID)
	if err != nil || report.PublishedSandbox != 1 || report.Impressions != 100 || report.Clicks != 10 || report.Conversions != 2 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}

func TestCompanySocialExternalPublishRemainsBlocked(t *testing.T) {
	store, _ := NewCompanyStore("")
	company, err := store.Create(Company{Name: "External test", OrganizationID: "org_social"})
	if err != nil {
		t.Fatal(err)
	}
	company, err = store.CreateSocialDraft(company.ID, CompanySocialDraft{Provider: "instagram", Title: "external", Body: "hello", Mode: "external"})
	if err != nil {
		t.Fatal(err)
	}
	draftID := company.SocialDrafts[0].ID
	_, err = store.ApproveSocialDraft(company.ID, draftID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.PublishSocialDraft(company.ID, draftID); !errors.Is(err, ErrCompanySocialExternalUnavailable) {
		t.Fatalf("expected external publishing blocker, got %v", err)
	}
}
