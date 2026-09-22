package agent

import (
	"errors"
	"testing"
)

func TestCompanyGrowthLifecycleAndApprovals(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org_local", Name: "Growth Smoke", Budget: CompanyBudget{Currency: "BRL", MonthlyLimitCents: 100000}})
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.AddCampaign(company.ID, CompanyCampaign{Name: "Lançamento", Channel: "social", Objective: "Gerar leads", DailyBudgetCents: 500})
	if err != nil || len(created.Campaigns) != 1 {
		t.Fatalf("campaign create failed: %v", err)
	}
	campaignID := created.Campaigns[0].ID
	if _, err := store.LaunchCampaign(company.ID, campaignID); !errors.Is(err, ErrCompanyApprovalRequiredForExternal) {
		t.Fatalf("expected campaign approval, got %v", err)
	}
	if _, err := store.ApproveCampaign(company.ID, campaignID); err != nil {
		t.Fatal(err)
	}
	created, err = store.LaunchCampaign(company.ID, campaignID)
	if err != nil || created.Campaigns[0].Status != "active" {
		t.Fatalf("campaign launch failed: %v", err)
	}

	created, err = store.AddAffiliateProgram(company.ID, CompanyAffiliateProgram{Name: "Parceiros", Network: "sandbox-network", CommissionBps: 1000})
	if err != nil {
		t.Fatal(err)
	}
	programID := created.AffiliatePrograms[0].ID
	if _, err := store.AddAffiliateLink(company.ID, CompanyAffiliateLink{ProgramID: programID, Destination: "https://shop.example.test/product"}); !errors.Is(err, ErrCompanyApprovalRequiredForExternal) {
		t.Fatalf("expected affiliate approval, got %v", err)
	}
	if _, err := store.ApproveAffiliateProgram(company.ID, programID); err != nil {
		t.Fatal(err)
	}
	created, err = store.AddAffiliateLink(company.ID, CompanyAffiliateLink{ProgramID: programID, Destination: "https://shop.example.test/product"})
	if err != nil || len(created.AffiliateLinks) != 1 {
		t.Fatalf("affiliate link failed: %v", err)
	}
	if _, err := store.RecordAffiliateConversion(company.ID, created.AffiliateLinks[0].ID, 1200); err != nil {
		t.Fatal(err)
	}

	created, err = store.AddProduct(company.ID, CompanyProduct{SKU: "SKU-001", Name: "Produto sandbox", Supplier: "Fornecedor sandbox", CostCents: 500, PriceCents: 1200, Inventory: 3})
	if err != nil {
		t.Fatal(err)
	}
	productID := created.Products[0].ID
	created, err = store.CreateOrder(company.ID, CompanyOrder{ProductID: productID, CustomerRef: "customer-test", Quantity: 2})
	if err != nil || len(created.Orders) != 1 || created.Orders[0].Status != "pending_approval" {
		t.Fatalf("order create failed: %v", err)
	}
	orderID := created.Orders[0].ID
	if _, err := store.FulfillOrder(company.ID, orderID, "TRACK-001"); !errors.Is(err, ErrCompanyApprovalRequiredForExternal) {
		t.Fatalf("expected order approval, got %v", err)
	}
	if _, err := store.ApproveOrder(company.ID, orderID); err != nil {
		t.Fatal(err)
	}
	created, err = store.FulfillOrder(company.ID, orderID, "TRACK-001")
	if err != nil || created.Orders[0].Status != "fulfilled" || created.Products[0].Inventory != 1 {
		t.Fatalf("fulfillment failed: %v", err)
	}
	if _, err := store.FulfillOrder(company.ID, orderID, "TRACK-001"); err != nil {
		t.Fatalf("fulfillment should be idempotent: %v", err)
	}
	if _, err := store.AddProduct(company.ID, CompanyProduct{SKU: "sku-001", Name: "Duplicado", PriceCents: 1200, Inventory: 1}); err == nil {
		t.Fatal("expected duplicate SKU rejection")
	}

	report, err := store.GrowthReport(company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if report.CampaignsActive != 1 || report.AffiliateConversions != 1 || report.FulfilledOrders != 1 || report.RevenueCents != 3600 {
		t.Fatalf("unexpected growth report: %+v", report)
	}
}

func TestGrowthExternalURLPolicy(t *testing.T) {
	store, err := NewCompanyStore("")
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.Create(Company{OrganizationID: "org_local", Name: "URL Policy"})
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.AddAffiliateProgram(company.ID, CompanyAffiliateProgram{Name: "Program", Network: "sandbox", CommissionBps: 100})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ApproveAffiliateProgram(company.ID, created.AffiliatePrograms[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddAffiliateLink(company.ID, CompanyAffiliateLink{ProgramID: created.AffiliatePrograms[0].ID, Destination: "http://external.example.test/product"}); !errors.Is(err, ErrCompanyInvalidExternalURL) {
		t.Fatalf("expected HTTPS policy, got %v", err)
	}
}
