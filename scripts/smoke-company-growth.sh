#!/usr/bin/env bash
set -euo pipefail
BASE="${AGENT_BASE_URL:-http://127.0.0.1:3001/api/agent/v1}"
JSON='Content-Type: application/json'
create=$(curl -fsS -X POST "$BASE/companies" -H "$JSON" -d '{"name":"Growth Smoke Company","mission":"Validar automação comercial em sandbox","currency":"BRL","budget":{"currency":"BRL","monthly_limit_cents":100000}}')
company_id=$(jq -r '.id' <<<"$create")
test -n "$company_id" -a "$company_id" != null
campaign=$(curl -fsS -X POST "$BASE/companies/$company_id/campaigns" -H "$JSON" -d '{"name":"Campanha Smoke","channel":"social","objective":"Gerar leads","daily_budget_cents":0}')
campaign_id=$(jq -r '.campaigns[-1].id' <<<"$campaign")
test -n "$campaign_id" -a "$campaign_id" != null
status=$(curl -sS -o /tmp/company-growth-campaign-approval.json -w '%{http_code}' -X POST "$BASE/companies/$company_id/campaigns/$campaign_id/launch" -H "$JSON" -d '{}')
test "$status" = 409
curl -fsS -X POST "$BASE/companies/$company_id/campaigns/$campaign_id/approve" -H "$JSON" -d '{}' >/dev/null
curl -fsS -X POST "$BASE/companies/$company_id/campaigns/$campaign_id/launch" -H "$JSON" -d '{}' | jq -e '.campaigns[-1].status == "active"' >/dev/null
program=$(curl -fsS -X POST "$BASE/companies/$company_id/affiliate-programs" -H "$JSON" -d '{"name":"Programa Smoke","network":"sandbox-network","commission_bps":1000}')
program_id=$(jq -r '.affiliate_programs[-1].id' <<<"$program")
curl -fsS -X POST "$BASE/companies/$company_id/affiliate-programs/$program_id/approve" -H "$JSON" -d '{}' >/dev/null
link=$(curl -fsS -X POST "$BASE/companies/$company_id/affiliate-links" -H "$JSON" -d "{\"program_id\":\"$program_id\",\"destination\":\"https://shop.example.test/product\"}")
link_id=$(jq -r '.affiliate_links[-1].id' <<<"$link")
curl -fsS -X POST "$BASE/companies/$company_id/affiliate-links/$link_id/conversion" -H "$JSON" -d '{"revenue_cents":1200}' >/dev/null
product=$(curl -fsS -X POST "$BASE/companies/$company_id/products" -H "$JSON" -d '{"sku":"SKU-SMOKE","name":"Produto Smoke","supplier":"Fornecedor sandbox","cost_cents":500,"price_cents":1200,"inventory":10}')
product_id=$(jq -r '.products[-1].id' <<<"$product")
order=$(curl -fsS -X POST "$BASE/companies/$company_id/orders" -H "$JSON" -d "{\"product_id\":\"$product_id\",\"customer_ref\":\"customer-smoke\",\"quantity\":1}")
order_id=$(jq -r '.orders[-1].id' <<<"$order")
status=$(curl -sS -o /tmp/company-growth-order-approval.json -w '%{http_code}' -X POST "$BASE/companies/$company_id/orders/$order_id/fulfill" -H "$JSON" -d '{"tracking_code":"PREMATURE"}')
test "$status" = 409
curl -fsS -X POST "$BASE/companies/$company_id/orders/$order_id/approve" -H "$JSON" -d '{}' >/dev/null
curl -fsS -X POST "$BASE/companies/$company_id/orders/$order_id/fulfill" -H "$JSON" -d '{"tracking_code":"SANDBOX-001"}' | jq -e '.orders[-1].status == "fulfilled"' >/dev/null
report=$(curl -fsS "$BASE/companies/$company_id/growth/report")
jq -e '.campaigns_active == 1 and .affiliate_conversions == 1 and .fulfilled_orders == 1 and .revenue_cents == 2400' <<<"$report" >/dev/null
printf 'COMPANY_GROWTH_SMOKE=PASS company=%s campaign=%s program=%s product=%s order=%s\n' "$company_id" "$campaign_id" "$program_id" "$product_id" "$order_id"
