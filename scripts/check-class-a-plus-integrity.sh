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
	  "internal/agent/company_test.go"
	  "internal/agent/builder.go"
	  "internal/agent/builder_test.go"
	  "internal/agent/swarm.go"
	  "internal/agent/traces.go"
	  "internal/agent/devices.go"
	  "internal/agent/traces_test.go"
	  "internal/agent/evaluation.go"
	  "internal/agent/browser_helper.py"
	  "internal/agent/connectors.go"
	  "internal/agent/connectors_test.go"
	  "internal/agent/mcp_remote.go"
	  "internal/agent/mcp_remote_test.go"
	  "internal/agent/mcp.go"
	  "internal/agent/plugin_lifecycle_test.go"
	  "internal/grok/client.go"
	  "internal/grok/live.go"
	  "internal/multillm/router.go"
	  "internal/multillm/registry.go"
	  "server/agent_routes.go"
	  "server/grok_routes.go"
	  "server/plugin_routes.go"
	  "server/plugin_routes_test.go"
	  "server/builder_scope_test.go"
	  "server/p0_scope_test.go"
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
grep -q 'connectorPathMatches' internal/agent/connectors.go
grep -q 'validRemoteMCPHeaderName' internal/agent/mcp_remote.go
grep -q 'GetForOrganization' internal/agent/builder.go
grep -q 'CreateRequest' internal/agent/company.go
grep -q 'ErrCompanyAgentBudgetExceeded' internal/agent/company_agents.go
grep -q 'PlanForOrganization' internal/agent/swarm.go
grep -q 'ListForOrganization' internal/agent/traces.go
grep -q 'HeartbeatForOrganization' internal/agent/devices.go
grep -q 'DecideApprovalForActorCAS' internal/agent/runtime.go
grep -q 'requireApprovalApprover' server/agent_routes.go
grep -q 'nonce' app/ui/app/src/components/AgenticConsole.tsx
grep -q 'type CompanyApproval struct' internal/agent/company.go
grep -q 'DecideApproval' server/company_approval_routes.go
grep -q 'PendingApproval' internal/agent/company.go
grep -q 'CompanyCampaignApprovalHTTPUsesNonceAndOrganization' server/company_approval_test.go
grep -q 'RecordSpendRequest' internal/agent/company.go
grep -q 'decideCompanyApprovalByID' server/company_approval_routes.go
grep -q 'CompanyApprovalQueue' app/ui/app/src/components/CompanyApprovalQueue.tsx
grep -q 'func RedactValue' internal/agent/secrets.go
grep -q 'TestRuntimeRedactsStepResultsEventsTracesAndPersistence' internal/agent/runtime_test.go
grep -q 'redactMissionForPersistence' internal/agent/store.go
grep -q 'remote MCP destination connected to a private address' internal/agent/mcp_remote.go
grep -q 'TestRemoteMCPDialRejectsPrivateActualAddress' internal/agent/mcp_remote_test.go
grep -q 'MCP command must be an absolute executable path' internal/agent/mcp.go
grep -q 'TestMCPPayloadLimitAndCancellationRestart' internal/agent/mcp_test.go
grep -q 'configureMCPProcess' internal/agent/mcp_process_unix.go
grep -q 'media destination connected to a private address' internal/agent/media.go
grep -q 'validateMediaMagic' internal/agent/media.go
grep -q 'TestMediaMaterializeRejectsRedirectAndInvalidMagic' internal/agent/media_test.go
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
