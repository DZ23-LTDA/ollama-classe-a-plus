# Changelog — Ollama Classe A+

Este arquivo registra as entregas públicas da distribuição `DZ23-LTDA/ollama-classe-a-plus`. O projeto mantém a atribuição e a licença do Ollama upstream; os recursos agentic específicos estão descritos com seus limites no [guia Classe A+](docs/CLASS_A_PLUS_GUIDE.md).

## Unreleased — rodada de paridade observável

- Árvore de produto completa em [`docs/agentic/PRODUCT_TREE.md`](docs/agentic/PRODUCT_TREE.md), separando superfície observável, estado atual e alvo unificado.
- Matriz de paridade em [`docs/agentic/PARITY_MATRIX.md`](docs/agentic/PARITY_MATRIX.md), com gates para Manus, Claude, Codex, OmniRoute e demais famílias de harness.
- Shell desktop com rotas reais de Projetos, Biblioteca, Agendado, Habilidades, Plugins e Tarefas, além de navegação lateral Classe A+.
- Agentic Control Center na Settings com catálogo de modelos e configuração sanitizada; o estado sem backend permanece explícito.
- Endpoint `GET /api/agent/v1/config/safe` sem tokens, caminhos privados ou valores sensíveis.
- Preset OmniRoute local, opt-in HTTP loopback protegido e smoke tests de forwarding/bearer server-side.
- Screenshots reais das novas rotas capturadas com Chromium e notas de proveniência atualizadas.
- Menu lateral Classe A+ aberto por padrão, com smoke test Chromium de rotas, links, ações primárias e estados vazios.
- `UPSTREAM_BASE_COMMIT`, `UPSTREAM_POLICY.md`, guardrail de integridade e workflow CI para impedir perda silenciosa das superfícies agentic.

## [0.1.0] — Preview público

### Incluído

- Runtime agentic em Go com missões persistentes, planejamento, execução, observação, recovery, artifacts e approvals.
- Ferramentas com sandbox, contenção de workspace, políticas de SSRF/path e trilhas de auditoria.
- Browser Operator via Playwright, MCP stdio, memória, projetos, skills e scheduler/webhooks.
- UI Agentic Console para missões, métricas, multiagente, pesquisa e approvals.
- Base de companion desktop para Linux, macOS e Windows e cliente Expo mobile com cache/outbox offline.
- Conectores allowlisted, providers Claude/Codex, autenticação, RBAC, OAuth/OIDC/SAML/MFA, colaboração e device pairing.
- Filas, DLQ/replay, PostgreSQL/RLS, Redis, OTLP, WebSocket/mTLS e exportadores/builders.
- DeploymentManager com providers Vercel, Netlify e generic, approval obrigatório e controles de segurança de publicação.
- Documentação pública em português, arquitetura, API, integrações, roadmap, política de segurança e screenshots reais.

### Estado honesto do preview

Adapters, contratos, testes e documentação não significam que contas externas, credenciais, instaladores assinados, publicação em lojas, modelos multimídia ou testes físicos estejam disponíveis neste ambiente. Deploy real, SSO contra um IdP real, GPU/modelos locais e distribuição de companions permanecem validações dependentes de ambiente. Os mockups em `docs/images/mockups/` são conceitos; as limitações das screenshots reais estão em `docs/images/screens/SCREEN_CAPTURE_NOTES.md`.

### Verificação

Os gates comprovados desta revisão incluem testes focados do runtime/server, build Go, build/typecheck da UI e typecheck mobile em suas fases correspondentes. O workflow público [`dz23-agentic-quality`](.github/workflows/dz23-agentic-quality.yaml) executa os gates de qualidade e a integração distribuída no GitHub Actions.

[0.1.0]: https://github.com/DZ23-LTDA/ollama-classe-a-plus/releases/tag/v0.1.0

## Unreleased — fluxo vertical funcional e HarnessRouter — 2026-09-22

- CRUD real tenant-aware de Projetos e Agendado, listagem de missões e artifacts na Biblioteca, com ações server-side e confirmação para exclusão.
- Agentic Console com seleção de Ollama local, Claude, Codex, OmniRoute e projeto persistente; Nova tarefa cria missão real e mantém timeline/approvals.
- Home do shell com composer, recomendações e atalhos funcionais para slides, site, design, jogos e missões; fallback local-first não bloqueia a aplicação quando Settings está offline.
- Catalogação funcional de connectors, MCP stdio, skills e CLIs com normalização de respostas nulas e nenhuma exposição de secret.
- Endpoint JSON `GET /api/agent/v1/metrics` registrado para o Console e smoke E2E live cobrindo home, CRUD, Plugins, Skills e criação de missão.
- Adapter HarnessRouter por OpenAI Responses-compatible com `harness_id` server-side, teste de preservação de metadata e preset [`examples/dz23-harnessrouter.json`](examples/dz23-harnessrouter.json); a execução real continua dependente de instância, chave e harness instalados.
- Manual, matriz de paridade, roadmap, checkpoint de missão e screenshots atualizados.

## Unreleased — Company OS e Desktop Commander Remote MCP — 2026-09-22

- Company OS persistente tenant-aware em `internal/agent/company.go`, com identidade, posicionamento, modelo de negócio, departamentos virtuais, roadmap, metas/KPIs, backlog priorizado, ciclos diários/semanais, relatório operacional, budget, approvals, anomalias e pausa automática.
- API `/api/agent/v1/companies` e rota `/company` com criação, seleção, atualização, roadmap, metas, backlog, ciclos ligados ao scheduler, pausa/retomada e registro de gasto.
- Worker de schedules passou a respeitar `company://<id>` e não cria novas missões para uma empresa pausada.
- Adapter `RemoteMCPManager` para Streamable HTTP com HTTPS fora de loopback, timeout, allowlist, bearer opcional por variável server-side e tool `mcp.remote.call` com approval.
- Presets sem segredos para Desktop Commander local stdio e Remote MCP oficial em `examples/dz23-desktop-commander-mcp.json` e `examples/dz23-desktop-commander-remote.json`.
- Documentação pública de Company OS, Desktop Commander, integração, API, matriz de paridade e estado honesto atualizada.
- Verificações desta rodada: testes `internal/agent`, `server` e `internal/multillm`, testes JSON dos presets e build Vite/TypeScript aprovados. OAuth PKCE, pareamento, conta, agentes físicos, connectors de social/afiliados/dropshipping e operação empresarial real continuam dependências externas.
