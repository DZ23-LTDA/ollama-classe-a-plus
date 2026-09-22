#!/usr/bin/env bash
set -euo pipefail
BASE="${AGENT_BASE_URL:-http://127.0.0.1:3001/api/agent/v1}"
json='Content-Type: application/json'
create=$(curl -fsS -X POST "$BASE/companies" -H "$json" -d '{"name":"Smoke Company OS","mission":"Validar fluxo vertical","business_model":"SaaS","currency":"BRL","budget":{"currency":"BRL","monthly_limit_cents":10000,"approval_threshold_cents":1000}}')
id=$(jq -r '.id' <<<"$create")
test -n "$id" -a "$id" != null
curl -fsS "$BASE/companies/$id" | jq -e '.name == "Smoke Company OS" and (.departments | length) == 7' >/dev/null
curl -fsS -X PATCH "$BASE/companies/$id" -H "$json" -d '{"positioning":"Local-first e seguro"}' | jq -e '.positioning == "Local-first e seguro"' >/dev/null
curl -fsS -X POST "$BASE/companies/$id/roadmap" -H "$json" -d '{"title":"Validar oferta","priority":10}' | jq -e '.roadmap | length == 1' >/dev/null
curl -fsS -X POST "$BASE/companies/$id/goals" -H "$json" -d '{"title":"Clientes","metric":"clientes_pagos","target":5}' | jq -e '.goals | length == 1' >/dev/null
curl -fsS -X POST "$BASE/companies/$id/backlog" -H "$json" -d '{"title":"Landing page","priority":10}' | jq -e '.backlog | length == 1' >/dev/null
curl -fsS -X POST "$BASE/companies/$id/cycles" -H "$json" -d '{"name":"Smoke cycle","objective":"Revisar o backlog","frequency":"daily","interval_seconds":3600}' | jq -e '.cycles | length == 1 and .[0].schedule_id != null' >/dev/null
if curl -sS -o /tmp/companyos-approval.json -w '%{http_code}' -X POST "$BASE/companies/$id/spend" -H "$json" -d '{"category":"ads","amount_cents":100,"approved":false}' | grep -qx '409'; then :; else cat /tmp/companyos-approval.json; exit 1; fi
curl -fsS -X POST "$BASE/companies/$id/spend" -H "$json" -d '{"category":"ads","amount_cents":100,"approved":true}' | jq -e '.budget.spent_cents == 100' >/dev/null
curl -fsS "$BASE/companies/$id/report" | jq -e '.open_backlog == 1 and .goals_on_track == 1 and .enabled_cycles == 1' >/dev/null
curl -fsS -X POST "$BASE/companies/$id/pause" -H "$json" -d '{"reason":"smoke pause"}' | jq -e '.status == "paused" and .risk.paused == true' >/dev/null
curl -fsS "$BASE/companies/$id" | jq -e '.status == "paused"' >/dev/null
curl -fsS -X POST "$BASE/companies/$id/resume" -H "$json" -d '{}' | jq -e '.status == "active" and .risk.paused == false' >/dev/null
printf 'COMPANY_OS_SMOKE=PASS id=%s\n' "$id"
