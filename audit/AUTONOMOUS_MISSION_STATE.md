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
