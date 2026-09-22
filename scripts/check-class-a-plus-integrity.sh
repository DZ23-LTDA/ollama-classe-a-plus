#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

required_files=(
  "UPSTREAM_BASE_COMMIT"
  "docs/agentic/PRODUCT_TREE.md"
  "docs/agentic/PARITY_MATRIX.md"
	  "docs/agentic/COMPANY_OS.md"
	  "docs/agentic/DESKTOP_COMMANDER_REMOTE.md"
	  "docs/agentic/COMPOSIO.md"
	  "docs/agentic/XAI_GROK.md"
	  "docs/agentic/EVALUATION.md"
	  "examples/dz23-composio-connect.json"
	  "examples/dz23-xai.json"
  "internal/agent/runtime.go"
  "internal/agent/company.go"
	  "internal/agent/company_growth.go"
	  "internal/agent/company_social.go"
	  "internal/agent/company_agents.go"
	  "internal/agent/evaluation.go"
	  "internal/agent/mcp_remote.go"
	  "internal/agent/plugin_lifecycle_test.go"
	  "internal/grok/client.go"
	  "internal/grok/live.go"
	  "internal/multillm/router.go"
	  "internal/multillm/registry.go"
	  "server/agent_routes.go"
	  "server/grok_routes.go"
	  "server/plugin_routes.go"
	  "server/company_routes.go"
  "server/company_growth_routes.go"
  "app/ui/app/src/components/AppSidebar.tsx"
  "app/ui/app/src/components/AgenticControlCenter.tsx"
  "app/ui/app/src/routes/agentic.tsx"
  "app/ui/app/src/routes/projects.tsx"
  "app/ui/app/src/routes/library.tsx"
  "app/ui/app/src/routes/scheduled.tsx"
  "app/ui/app/src/routes/skills.tsx"
  "app/ui/app/src/routes/plugins.tsx"
  "app/ui/app/src/routes/tasks.tsx"
  "app/ui/app/src/routes/company.tsx"
	  "app/ui/app/src/components/CompanyGrowthPanel.tsx"
	  "app/ui/app/src/components/CompanyOperationsPanel.tsx"
  "scripts/smoke-company-growth.sh"
  "scripts/smoke-builder.sh"
)

for file in "${required_files[@]}"; do
  test -s "$file" || { echo "missing required Class A+ surface: $file" >&2; exit 1; }
done

grep -q '^policy=manual-review-only$' UPSTREAM_BASE_COMMIT
grep -q 'allow_insecure_loopback' internal/multillm/registry.go
grep -q 'safeConfig' server/agent_routes.go
grep -q 'Nova tarefa' app/ui/app/src/components/AppSidebar.tsx
grep -q 'Agentic Control Center' app/ui/app/src/components/AgenticControlCenter.tsx
grep -q 'RemoteMCP' internal/agent/runtime.go
grep -q 'mcp.remote.call' internal/agent/mcp_remote.go
grep -q 'HeadersEnv' internal/agent/mcp_remote.go
grep -q 'api.x.ai/v1' examples/dz23-xai.json
grep -q 'Growth OS' app/ui/app/src/components/CompanyGrowthPanel.tsx
grep -q 'addCompanyCampaign' server/company_growth_routes.go
grep -q 'grokResponses' server/grok_routes.go
grep -q 'SetConnectorEnabled' internal/agent/runtime.go
grep -q 'Capabilities' internal/agent/runtime.go
grep -q 'Evaluation' internal/agent/evaluation.go
for route in projects library scheduled skills plugins tasks company; do
  grep -q "routes/${route}" app/ui/app/src/routeTree.gen.ts || {
    echo "route tree is missing /${route}" >&2
    exit 1
  }
done

if git ls-files | grep -E '(^|/)(\.env|.*\.key|.*\.pem|node_modules/)' >/dev/null; then
  echo "tracked secret or generated dependency path detected" >&2
  exit 1
fi

git diff --check
printf '%s\n' "Class A+ integrity guard: PASS"
