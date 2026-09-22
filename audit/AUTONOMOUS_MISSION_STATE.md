# Estado da missão autônoma — Ollama DZ23 Agentic Platform

```yaml
mission_id: dz23-agentic-platform-2026-09-21
objective: Evoluir o Ollama DZ23 para uma plataforma agentic local-first com execução segura, ferramentas, memória, artefatos, automações, integrações e superfícies Desktop/Mobile.
state: RECOVERING
iteration: 1
started_at: 2026-09-21
last_progress_at: 2026-09-21
branch: feat/dz23-claude-codex-desktop
base_commit: 4a3dd89ed9224f2d6cd87579434ca6798334fa4a

acceptance:
  - O núcleo de missão possui estados explícitos, persistência, eventos e recuperação.
  - Toda ferramenta possui contrato, limites, autorização e resultado auditável.
  - Terminal e execução de código são isolados por política; não há shell arbitrário por padrão.
  - Memória, projetos, skills, MCP, artefatos e jobs têm fronteiras versionadas.
  - Browser/computer-use, integrações externas, mídia, Web/Desktop/Mobile têm adapters testáveis.
  - Fluxos críticos têm testes unitários, contrato, integração, segurança e smoke real.
  - Nenhuma capacidade é marcada como completa sem prova observável.

completed:
  - Auditoria do commit 4a3dd89e e confirmação da base multi-provider.
  - Ponte Anthropic Messages para providers OpenAI-compatible.
  - Inventário externo para Claude Desktop e Codex Desktop.
  - Testes focados de multillm, proxy e launch.

current_task: Arquitetura e núcleo vertical de missões.
pending:
  - Runtime de missão com planner, executor, observer e recovery.
  - Store durável, eventos e idempotência.
  - Tool registry, approvals, sandbox e artifacts.
  - Memória/projetos/skills/MCP.
  - Browser/computer-use e integrações.
  - Mídia, builder, jogos, slides, dados e mobile.
  - Auditorias independentes, build e publicação.

blockers:
  - GitHub push bloqueado por 403 para a identidade dz23trading-collab.
  - Build nativo completo depende das toolchains de cada plataforma.
  - Integrações externas reais dependem de credenciais e autorização do usuário.

risks:
  - Execução de shell, browser, desktop e conectores podem produzir efeitos externos; exigir aprovação e allowlists.
  - Não implementar browser/computer-use falso baseado apenas em respostas do modelo.
  - Não persistir segredos em missão, memória, logs ou artefatos.

next_action: Criar contratos agentic e implementar uma missão vertical persistente com ferramenta segura de filesystem.
```

## Regra de retomada

Antes de continuar, conferir este arquivo contra `git status`, o commit atual, os testes e os artefatos. Retomar pela primeira tarefa não concluída; não repetir a ponte Claude/Codex já validada.


## Adendo — fase agentic multimodal, builders, auth e colaboração — 2026-09-21

```yaml
state: VERIFIED_LOCAL_PHASE
completed:
  - auth: organizations, memberships, RBAC, revocable tokens, OAuth PKCE state, AES-GCM credential storage and refresh contract
  - jobs: persistent queue, retries, dead-letter queue, replay, trace spans and SSE events
  - media: HTTPS image/video/speech/transcription adapters plus deterministic WAV smoke fixture
  - builders: website/app/game/slides/dashboard templates, preview containment, ZIP export and local versioned publish
  - desktop: Linux implementation preserved, Darwin and Windows adapters compile cross-platform
  - collaboration: persistent comments, presence, snapshots and SSE stream
  - mobile: Expo SecureStore session, EAS profiles, Android/iOS identifiers and typecheck
proofs:
  - CGO_ENABLED=0 go test ./internal/agent -count=1: PASS
  - GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ./internal/agent: PASS
  - GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./internal/agent: PASS
  - apps/mobile-agentic npm ci && npm run typecheck: PASS
  - server gate: BLOCKED by existing upstream MLX symbols/toolchain, not by internal/agent tests
next_action: Resolve MLX build environment, then commit/review/push the verified phase.
```


## Adendo — fase 4 multiagente, pesquisa, devices e ingestão — 2026-09-21

```yaml
state: TESTING
implemented:
  - multiagent_orchestrator: specialist_roles, concurrency_budget, retries, cancellation, persistence and synthesis conflict detection
  - deep_research: multi-source HTTPS fetch, HTML extraction, cache, citations, hashes, robots policy and SSRF guard
  - device_pairing: one-time pairing, capability report, heartbeat, online/offline state and revocation
  - document_ingestion: txt/md/html/json/csv/pdf/docx/xlsx readers, chunking and provenance-backed memories
proofs:
  - CGO_ENABLED=0 go test ./internal/agent -count=1 after phase 4: PASS
  - focused swarm/research/device tests: PASS
remaining:
  - resolve full server build with CGO/MLX toolchain and run server tests
  - integrate production mTLS/WebSocket companion transport, OAuth provider adapters and distributed stores
next_action: run full server gate after build-essential installation, then commit phase 4 and package artifacts.
```


## Fechamento da fase 4 — 2026-09-21

```yaml
state: CANDIDATE_COMPLETED
proofs:
  - CGO_ENABLED=0 go test ./internal/agent -count=1: PASS
  - CGO_ENABLED=1 go test ./server -count=1: PASS
  - CGO_ENABLED=1 go build -o ollama-dz23-agentic-server .: PASS
  - frontend ./node_modules/.bin/tsc --noEmit: PASS
  - git diff --check: PASS
features:
  - multiagent orchestration with seven roles, budgets, retries, cancellation and synthesis
  - deep research with citations, hashes, cache, HTML extraction, robots and SSRF policy
  - device pairing, capability report, heartbeat, offline state and revocation
  - safe PDF/DOCX/XLSX/document ingestion with chunks and provenance
  - Web Agentic Console controls for missions, orchestration and research
blockers:
  - production mTLS/WebSocket companion transport, OCR/vision providers, distributed stores, public hosting deploy, signed desktop/mobile releases and physical device validation remain
next_action: run final diff review, commit phase 4 and build verified ZIP; do not publish main.
```


## Adendo — fase 5 infraestrutura distribuída, companion seguro e mobile offline — 2026-09-21

```yaml
state: TESTING
implemented:
  - postgres_store: idempotent migrations, mission upsert and append-only event persistence
  - redis_queue: enqueue, claim, retry backoff, dead-letter, replay and worker integration
  - opentelemetry: optional OTLP HTTP exporter with HTTPS-by-default and local noop fallback
  - companion_transport: WebSocket handshake, device token, capabilities, heartbeat, TLS/mTLS policy and origin allowlist
  - mobile_offline: cached mission, queued actions and synchronization retry
  - ci_sbom: Go/web/mobile gates and CycloneDX artifact workflow
  - local_stack: PostgreSQL, Redis and OTEL Collector compose files
proofs:
  - CGO_ENABLED=0 go test ./internal/agent: PASS
  - CGO_ENABLED=1 go test ./server: PASS
  - apps/mobile-agentic npm ci && npm run typecheck: PASS
pending_production:
  - integration tests against real PostgreSQL/Redis/OTLP endpoints
  - certificate rotation, RLS and tenant isolation review
  - remote push notifications, mobile conflict resolution and physical device tests
next_action: verify root build, diff, commit phase 5 and package ZIP; preserve unresolved external credential/hardware gates.
```


## Fechamento da fase 5 — 2026-09-21

```yaml
state: CANDIDATE_COMPLETED
commit: 7b0609d8e43414c006150bc4de9f5fbe9571b57a
proofs:
  - CGO_ENABLED=0 go test ./internal/agent -count=1: PASS
  - CGO_ENABLED=1 go test ./server -count=1: PASS
  - apps/mobile-agentic npm ci && npm run typecheck: PASS
  - CGO_ENABLED=1 go build -o ollama-dz23-agentic-phase5 .: PASS
  - zipinfo -t phase5 archive: PASS
features:
  - PostgreSQL store and Redis queue adapters with local fallback
  - OTLP OpenTelemetry exporter with HTTPS-by-default and noop fallback
  - TLS/mTLS policy and WebSocket companion handshake/heartbeat
  - mobile cached mission and offline action queue
  - CI quality workflow, CycloneDX SBOM and local infra compose stack
external:
  - push remains blocked by GitHub HTTP 403 for dz23trading-collab
remaining:
  - real PostgreSQL/Redis/OTLP integration tests, certificate rotation, RLS, remote push, physical devices, public deploy adapters, OCR/model providers and signed releases
next_action: proceed to the production-adapter phase only after external credentials, certificates and test infrastructure are available.
```


## Fechamento da fase 7 — 2026-09-21

```yaml
state: CANDIDATE_COMPLETED
features:
  - saml_sp: crewjam metadata validation, signed AuthnRequest, one-time RelayState, ACS claims and tenant provisioning
  - oauth_connectors: tenant-aware encrypted credential resolution for connector.http
  - builder_history: rich component fields, validation, persistent undo/redo and API endpoints
  - postgres_rls: FORCE ROW LEVEL SECURITY and tenant policies excluding blank organization records
  - companion_tls: TLS 1.3/mTLS listener configuration with per-handshake certificate reload
proofs:
  - CGO_ENABLED=0 go test ./internal/agent -count=1: PASS
  - CGO_ENABLED=1 go test ./server ./cmd/launch ./internal/multillm -count=1: PASS
  - CGO_ENABLED=1 go build -o ollama-dz23-phase7-bin .: PASS
  - apps/mobile-agentic npm ci && npm run typecheck: PASS
  - JSON manifests and git diff --check: PASS
remaining_external:
  - SAML end-to-end IdP and certificate fixtures
  - real PostgreSQL/Redis/OTLP execution outside CI and production RLS migration review
  - physical companion/mobile tests, push credentials, signed installers, app-store distribution and public cloud deploy credentials
  - full editorial exporters, CRDT collaboration, local generative media models and complete hosting adapters
next_action: commit phase 7, create reproducible archive, then continue with deploy adapters and physical/infrastructure gates without publishing main.
```


## Fechamento da fase 8 — 2026-09-21

```yaml
state: CANDIDATE_COMPLETED
features:
  - deployments: Vercel, Netlify and generic hosting adapters
  - publish_security: workspace containment, symlink rejection, file/size limits, no redirects, HTTPS outside loopback
  - publish_approval: external deployment endpoint requires approved=true
proofs:
  - CGO_ENABLED=0 go test ./internal/agent -count=1: PASS
  - CGO_ENABLED=1 go test ./server ./cmd/launch ./internal/multillm -count=1: PASS
  - CGO_ENABLED=1 go build -o ollama-dz23-phase8-bin .: PASS
  - generic HTTPS deployment smoke and external HTTP rejection: PASS
remaining_external:
  - real Vercel/Netlify/AWS/Cloudflare accounts, project IDs and permissions
  - signed installers, physical devices, store distribution and full local media models
next_action: commit phase 8, package a reproducible archive and continue provider-specific production smoke tests only with operator credentials.
```

## Rodada 2026-09-22 — árvore de produto, shell desktop e OmniRoute
```yaml
state: CANDIDATE_COMPLETED
branch: feat/manus-parity-omniroute
features:
  - product_tree: Manus observable surface, current Classe A+ tree and unified target tree
  - parity_matrix: evidence states and acceptance journeys
  - desktop_shell: real routes for projects, library, scheduled, skills, plugins and tasks
  - settings_control_center: sanitized runtime/provider/approval/integration status
  - safe_config_endpoint: GET /api/agent/v1/config/safe without secrets or private paths
  - omniroute_loopback: explicit local HTTP opt-in plus preset and proxy smoke tests
  - screenshots: Chromium captures for shell routes and Settings
proofs:
  - CGO_ENABLED=0 go test ./internal/agent -count=1: PASS
  - CGO_ENABLED=1 go test ./server ./cmd/launch ./internal/multillm -count=1: PASS
  - UI npm run build: PASS
  - Chromium route capture: PASS for projects, library, scheduled, skills, plugins, tasks, agentic and settings
  - git diff --check: PASS
remaining_external:
  - OmniRoute instance, credentials, provider discovery and fallback smoke outside the fixture
  - CRUD persistence for new shell surfaces and complete provider settings editor
  - real browser/desktop/mobile devices, distributed staging, IdP, deploy accounts and signed releases
next_action: review diff, commit the feature branch, publish branch for review and continue the vertical flow Nova tarefa -> missão -> approval -> artifact.
```


## Checkpoint atual — fluxo vertical e HarnessRouter — 2026-09-22

```yaml
state: CANDIDATE_COMPLETED
branch: feat/manus-parity-omniroute
completed:
  - shell_home: composer, recommendations, shortcuts and local-first offline fallback
  - project_schedule_crud: tenant-aware create/list/update/delete with live API
  - mission_console: provider selector for Ollama, Claude, Codex, OmniRoute and project association
  - catalogs: connectors, MCP, skills and CLI status with null-safe frontend normalization
  - observability: JSON metrics route registered for Agentic Console
  - harnessrouter_adapter: OpenAI Responses-compatible provider with server-side harness_id metadata injection
  - public_docs: README, manual, parity matrix, roadmap, changelog and screenshots updated
proofs:
  - go test ./internal/agent ./server -count=1: PASS
  - UI npm run build: PASS
  - live CRUD smoke for projects and schedules against 127.0.0.1:3001: PASS
  - live mission/list/events smoke: PASS
  - Chromium capture for home, projects, library, scheduled, skills, plugins, tasks, agentic and settings: PASS
  - node app/ui/app/scripts/smoke-shell.mjs against Vite and local server: PASS
  - HarnessRouter metadata proxy regression test: pending final Go gate
remaining_external:
  - HarnessRouter instance, provider key and installed harnesses for streaming/follow-up/cancel/artifact validation
  - builder drag-and-drop and CRDT collaboration
  - distributed PostgreSQL/Redis/OTLP, IdP, real deploy accounts and physical desktop/mobile tests
next_action: run focused multillm/server gates, review diff, commit and publish the feature branch without changing main.
```


## Checkpoint atual — Company OS e Desktop Commander Remote MCP — 2026-09-22

```yaml
state: CANDIDATE_COMPLETED
branch: feat/manus-parity-omniroute
completed:
  - company_os_store: tenant-aware identity, departments, roadmap, goals, backlog, cycles, budget and risk pause
  - company_os_api: CRUD identity, report, roadmap, goals, backlog, cycles, pause/resume, anomaly and spend endpoints
  - company_os_ui: /company functional workspace with creation, KPI/backlog/cycle forms and guardrails
  - company_scheduler_guard: company:// workspaces skip mission creation while paused
  - remote_mcp_adapter: Streamable HTTP, HTTPS policy, timeout, allowlist, optional server-side bearer and approval tool
  - desktop_commander_presets: local stdio and official remote endpoint examples without secrets
  - public_docs: Company OS, Remote MCP, API, integrations, parity matrix, roadmap and changelog
proofs:
  - go test ./internal/agent ./server ./internal/multillm -count=1: PASS
  - remote MCP httptest JSON/SSE, bearer and URL policy tests: PASS
  - CompanyStore lifecycle, persistence, budget pause and anomaly tests: PASS
  - JSON preset validation: PASS
  - UI npm run build: PASS
remaining_external:
  - Desktop Commander OAuth PKCE, account, device pairing, physical agent and revocation flow
  - real Remote MCP call against an authorized paired test device
  - CRM, social, affiliate, ecommerce, advertising, logistics and analytics connectors
  - distributed Company OS storage/RLS, IdP, deploy accounts and physical desktop/mobile tests
next_action: review diff, run integrity and full gates, update screenshots if needed, commit and publish the feature branch without changing main
```


## Verificação final da rodada — 2026-09-22

```yaml
state: CANDIDATE_COMPLETED
proofs:
  - Class A+ integrity guard: PASS
  - CGO_ENABLED=0 go test ./internal/agent -count=1: PASS
  - CGO_ENABLED=1 go test ./server ./cmd/launch ./internal/multillm -count=1: PASS
  - CGO_ENABLED=1 go build: PASS
  - UI npm run build: PASS
  - Company OS live API smoke: PASS
  - Chromium shell smoke including /company and MCP catalog: PASS
  - Company OS screenshot capture and visual inspection: PASS
  - git diff --check: PASS
external_blockers_unchanged:
  - OAuth PKCE, device pairing and physical Desktop Commander agent
  - real social, CRM, affiliate, ecommerce, ads, logistics and analytics providers
  - distributed Company OS RLS, IdP, deployment accounts and signed devices
next_action: commit and push the verified feature branch, then update the public PR; do not merge into main automatically
```


## Verificação adicional — jornadas verticais Growth/Builder/providers — 2026-09-22

```yaml
state: CANDIDATE_COMPLETED
branch: feat/manus-parity-omniroute
completed:
  - growth_os_sandbox: campaigns, approvals, launch/pause, affiliate programs/links/conversions, products, orders, inventory and fulfillment
  - company_growth_ui: Growth OS panel embedded in /company with local metrics and explicit sandbox labels
  - builder_vertical: create, visual update, preview/hash artifact, export, local publish and external deploy approval gate
  - provider_sessions: OpenAI-compatible streaming translation with gateway auth and server-side provider bearer
  - remote_mcp_evidence: official setup source recorded; adapter tests remain green
proofs:
  - go test ./internal/agent ./server -count=1: PASS
  - focused multi-provider streaming/OmniRoute/HarnessRouter tests: PASS
  - focused Remote MCP tests: PASS
  - scripts/smoke-company-growth.sh against 127.0.0.1:3001: PASS
  - scripts/smoke-builder.sh against 127.0.0.1:3001: PASS
  - UI npm run build: PASS
  - Chromium capture /company with Growth OS: PASS
external_blockers:
  - Desktop Commander OAuth PKCE, account, device pairing, physical agent, remote call and revocation; no test token/device is provisioned
  - real social/CRM/affiliate/ecommerce/ads/logistics/analytics connectors and compliance staging
  - real provider sessions for Claude/Codex/OmniRoute/HarnessRouter with operator credentials
  - distributed PostgreSQL/Redis/OTLP, IdP, deployment accounts and physical desktop/mobile tests
next_action: run full integrity/backend/UI gates, stage the complete diff, commit and update the public PR
```


## Verificação adicional — Composio, xAI/Grok e social commerce — 2026-09-22

```yaml
state: CANDIDATE_COMPLETED
branch: feat/manus-parity-omniroute
completed:
  - composio_remote_mcp_preset: headers_env server-side, JSON-RPC allowlist, approval and local regression test
  - xai_responses_preset: HTTPS OpenAI-compatible provider, bearer server-side and local passthrough test
  - commercial_source_audit: official Composio, xAI, TikTok Shop, Instagram and Shopify sources recorded
  - parity_docs: README, integrations, parity matrix, changelog and roadmap updated
proofs:
  - gofmt and focused Remote MCP tests: PENDING_FINAL_GATE
  - focused multi-provider xAI Responses test: PENDING_FINAL_GATE
  - JSON preset validation: PENDING_FINAL_GATE
  - integrity/backend/UI gates: PENDING_FINAL_GATE
external_blockers:
  - Composio API key, OAuth connected accounts and per-tenant session provisioning
  - xAI API key/quota and provider tool validation; Grok Bot cloud product is not embedded
  - TikTok Shop Partner Center app, seller/creator/partner authorization, scopes and region-specific sandbox
  - Meta/Instagram, YouTube, WhatsApp, Shopify and other marketplace credentials, app review and webhooks
  - compliance, idempotency, DLP, approvals, returns/refunds, logistics, payments and reconciliation tests
next_action: run final gates, commit and push the public PR; do not claim external accounts are connected
```
