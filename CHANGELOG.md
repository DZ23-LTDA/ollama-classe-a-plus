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


## Unreleased — jornadas verticais operacionais — 2026-09-22

- Growth OS local no Company OS: campanhas com approval/launch/pause, programas e links de afiliados com destino HTTPS, conversões, catálogo de produtos, pedidos, approval, fulfillment sandbox, estoque e relatório agregado.
- Painel visual funcional na rota `/company` para criar e operar campanhas, programas, produtos e pedidos sem transformar estados externos em sucesso falso.
- Smoke live `scripts/smoke-company-growth.sh` com criação, bloqueio por ausência de approval, aprovação, execução sandbox, conversão, fulfillment e validação de métricas.
- Smoke live `scripts/smoke-builder.sh` com criação, editor visual, preview com hash, export, publicação local e bloqueio de deploy externo sem aprovação explícita.
- Teste de provider session streaming OpenAI-compatible com autenticação do gateway, bearer server-side e tradução de eventos para o protocolo nativo.
- Teste do adapter Desktop Commander Remote MCP e registro da evidência oficial de setup; OAuth, conta, pareamento, agente físico e revogação continuam pendentes por falta de credenciais/dispositivo autorizados.

A rodada não publica campanhas, não movimenta dinheiro, não envia pedidos, não conecta contas sociais/CRM/marketplaces e não declara providers externos como conectados. Os fluxos comprovados são locais e reversíveis.


## Unreleased — Composio, xAI/Grok e social commerce — 2026-09-22

- Remote MCP agora aceita `headers_env` server-side com validação de nomes, permitindo o preset oficial Composio Connect sem expor `x-consumer-api-key` na UI, memória ou logs.
- Preset [`examples/dz23-composio-connect.json`](examples/dz23-composio-connect.json) com endpoint HTTPS, métodos JSON-RPC allowlisted e approval via `mcp.remote.call`.
- Preset [`examples/dz23-xai.json`](examples/dz23-xai.json) para xAI Responses/chat com bearer server-side, e teste de passthrough de `input`/`tools` para `/v1/responses`.
- Runbooks [`COMPOSIO.md`](docs/agentic/COMPOSIO.md) e [`XAI_GROK.md`](docs/agentic/XAI_GROK.md), além de fontes oficiais registradas em [`audit/COMMERCIAL_INTEGRATIONS_SOURCES_2026-09-22.md`](audit/COMMERCIAL_INTEGRATIONS_SOURCES_2026-09-22.md).
- Matriz de paridade atualizada com Composio, xAI/Grok API e Social Commerce/TikTok Shop, separando adapter de conta, scopes, sandbox e operação real.

A rodada não conecta contas externas nem publica posts, anúncios, produtos ou pedidos. Composio, xAI, TikTok Shop, Instagram, Shopify e demais canais exigem credenciais, aprovação de app, scopes mínimos, testes de sandbox, compliance e validação por região.


## Unreleased — hardening interno e baseline de release — 2026-09-22

- Autenticação agentic agora exige bearer por padrão fora de loopback e não aceita `OLLAMA_AGENT_AUTH_REQUIRED=false` em bind externo.
- MCP stdio e Remote MCP exigem métodos allowlisted; Remote MCP usa timeout, bloqueio de destino privado, redirects sem mudança de origem e correlation ID JSON-RPC.
- Contenção de workspace rejeita componentes symlink e StepID fora do formato seguro antes de executar sandbox ou manipular arquivos.
- Approvals carregam organização, actor, policy, nonce e expiração; mutações de Growth validam tenant; pedidos usam `Idempotency-Key`; fulfillment repetido não reduz estoque novamente.
- Control Center e Settings normalizam catálogos nulos; outbox mobile remove bearer persistido, usa idempotency key e backoff com limite de tentativas.
- Teste obsoleto que rejeitava a CLI agentic foi corrigido para exigir sua disponibilidade, mantendo rejeição dos flags antigos.
- Gates aprovados: `CGO_ENABLED=1 go test ./... -count=1`, `CGO_ENABLED=1 go vet ./...`, integrity guard, build Go, build UI, Vitest 20/199 e typecheck mobile.

O hardening não transforma adapters externos em contas conectadas. Composio, xAI, Desktop Commander, redes sociais, marketplaces, PostgreSQL/RLS, Redis, OTLP, dispositivos físicos, instaladores assinados e lojas continuam dependências externas a validar.


### Incremento final de policy — 2026-09-22

Missões agora persistem `capabilities` e recebem somente `workspace:read`/`workspace:write` por default. Tools com scopes de browser, desktop, terminal, sandbox, MCP, connectors ou deploy são negadas até que o escopo seja concedido explicitamente; approvals e allowlists continuam independentes. O bypass de autenticação foi limitado ao FullPath exato do handshake `/devices/:id/connect`, em vez de liberar qualquer caminho terminado em `/connect`.


## Unreleased — capacidades operacionais e lifecycle — 2026-09-22

- Grok Live local-first em `internal/grok`: Responses API, streaming SSE, retries limitados, circuito de falha, status sanitizado e distinção entre fontes ao vivo e memória.
- Evaluation OS determinístico para coding, browser, tools, segurança, memória, planejamento e recuperação; provider router por capacidades, saúde, latência, custo, privacidade e qualidade histórica.
- Company Operations com agentes departamentais persistentes, supervisor, orçamento, SLA, memória, pausa/retomada e Social OS sandbox para contas OAuth pendentes, drafts, approvals, publicação simulada e métricas.
- Lifecycle server-side para connectors, MCP stdio, Remote MCP e skills: habilitar, desabilitar e remover com bloqueio de execução e catálogo sanitizado.
- Correção do deploy Builder: `approved:false` retorna `428 Precondition Required` antes de consultar providers ou configuração de deploy ausente.
- Smoke real aprovado para Growth OS, Builder e shell Chromium; build UI, Vitest, testes backend, integrity guard e contratos locais permanecem verdes.

A rodada continua sem conectar contas externas nem publicar posts, anúncios, produtos, pedidos ou deploys públicos. xAI, Composio, Desktop Commander, social commerce, marketplaces, PostgreSQL/RLS, Redis, OTLP, IdP, GPU, dispositivos físicos, assinatura e lojas exigem ambientes e credenciais autorizados.


## Unreleased — smoke de APIs e correção do Browser Operator — 2026-09-22

- Smoke seguro de catálogos provider com as chaves fornecidas, sem persistir ou imprimir valores secretos.
- Inferência curta em modelo gratuito do OpenRouter aprovada diretamente e pelo gateway local Classe A+ (`/v1/chat/completions`).
- Relatório público de prontidão em [`docs/agentic/READINESS_2026-09-22.md`](docs/agentic/READINESS_2026-09-22.md), distinguindo adapters, validações locais e blockers externos.
- Workflow agentic-quality atualizado para instalar Playwright/Chromium antes do teste real do Browser Operator.
- Erros do Browser Operator agora incluem stderr sanitizado para diagnóstico de dependência, sem expor credenciais.

As credenciais fornecidas no anexo foram tratadas como expostas por terem sido compartilhadas em texto aberto; devem ser revogadas e recriadas pelo operador. Nenhum provider externo é promovido a integração de produção apenas por responder a um health check.


## Unreleased — hardening de connectors, plugins e Browser Operator — 2026-09-22

- Connectors passaram a exigir escopo de organização em chamadas operacionais, respeitar fronteiras de segmento nos `path_prefixes` e rejeitar destinos privados resolvidos por DNS, com exceção explícita apenas para loopback configurado.
- Remote MCP e MCP stdio passaram a validar nomes de ambiente; Remote MCP também bloqueia headers HTTP de transporte e framing.
- Lifecycle global de connectors, MCP, Remote MCP e skills exige `owner`/`admin` no modo autenticado; o modo local permanece restrito ao listener loopback.
- Browser Operator agora usa fallback de Chromium quando um caminho configurado está ausente, preservando diagnóstico e validação de URL.
- Gates locais aprovados: integrity guard, suíte Go completa com CGO, `go vet`, build Go, Vitest 20/199, build UI, typecheck mobile e teste Browser Operator com caminho inválido.

A classificação permanece release candidate local-first. Integrações externas, ambientes distribuídos, dispositivos físicos, assinaturas, lojas e deploys reais continuam condicionados a credenciais, aprovação, sandbox e validação do operador.


## 2026-09-22 — P0 ownership e Company invariants

- Builder projects receberam `OrganizationID` e todas as operações HTTP de projeto passaram a usar lookup/listagem scoped, com teste cross-tenant `403` e ausência de mutação.
- Builder agora exige entry presente, rejeita symlink no preview/export e escapa nomes inseridos nos templates HTML.
- Company create passou a usar DTO allowlisted e a resetar campos server-managed; gasto de agente pausado ou acima do budget falha atomicamente.
- Gates locais do slice passaram; isso não encerra os P0 restantes nem transforma adapters externos em integrações conectadas.


## 2026-09-22 — P0 tenant isolation em orchestration, traces e devices

- Orchestration jobs receberam `organization_id`; plan/get/run/cancel server-side passaram a usar métodos scoped.
- Spans de missão e ferramenta carregam organização e a API de traces filtra por tenant.
- Devices e pairing passaram a aplicar ownership em listagem, heartbeat, revoke e consumo de código de pairing, com regressões HTTP cross-tenant `403` e sem mutação.
- Gates completos da slice passaram; companion físico, mTLS e os demais P0 continuam sem prova de produção.


## 2026-09-22 — mission approvals com policy, nonce e CAS

- Decisão de approval autenticada exige membership owner/admin; operator/viewer não podem aprovar.
- O endpoint usa nonce de uso único e versão corrente da missão, rejeitando replay e conflito de concorrência.
- Agentic Console e tipos de API foram atualizados para transportar nonce/policy.
- Company/Growth/Social approvals booleanas continuam explicitamente fora desta garantia até migrarem para decisões auditáveis.


## 2026-09-22 — Company OS approvals auditáveis

Campanhas, afiliados, pedidos e drafts sociais passaram a emitir `CompanyApproval` server-side com policy, nonce, expiração, actor, organização, razão e versão CAS. Os endpoints de aprovação não aceitam mais `{}` como autorização: localizam a decisão pendente e só projetam o recurso aprovado após validação. A rota de gasto com `approved` booleano permanece aberta para a próxima slice.


## 2026-09-22 — spend approvals e budget atômico

A rota HTTP de gasto deixou de aceitar `approved` do cliente. Solicitações que atravessam a política de budget viram approvals pendentes com nonce, expiração, actor, organização e CAS; a nova fila do Company OS decide e só então contabiliza o valor. O limite mensal continua pausando a Company sem débito parcial.


## 2026-09-22 — DLP de resultados e observabilidade

Resultados de tools, erros, eventos, traces e serializações JSON/Postgres passaram a usar redação recursiva de credenciais. A cobertura inclui campos estruturados sensíveis e tokens GitHub/OpenAI/OpenRouter/xAI/AWS/Slack/Bearer/PEM, com regressões de token injection. Egress de connectors/MCP permanece uma superfície distinta, pois credenciais operacionais não podem ser redigidas cegamente.


## 2026-09-22 — Remote MCP egress hardening

O cliente Remote MCP deixou de usar proxy ambiental, limita redirects ao mesmo origin e verifica o IP efetivamente conectado para rejeitar destinos privados e DNS rebinding. Foram adicionadas regressões para redirect same-origin e conexão TCP privada local.


## 2026-09-22 — MCP stdio lifecycle hardening

MCP stdio passou a usar executável absoluto não-symlink, cwd privado, ambiente mínimo, limites de argumentos/request/response/stderr, redaction de stderr, grupo de processo e restart/cleanup verificáveis. A implementação é explicitamente best-effort; seccomp/cgroups/rlimits e limites fortes de recursos continuam fora desta slice.


## 2026-09-22 — Media egress/download hardening

Downloads de mídia agora bloqueiam redirects e proxy ambiental, verificam o IP conectado, aplicam limites bounded e validam MIME/magic antes de materializar PNG/JPEG/WebP/MP4/WAV. Testes negativos cobrem redirects, MIME incompatível, conteúdo falso, overflow e destinos privados.


## 2026-09-22 — Connector egress parity

Connectors passaram a bloquear proxy ambiental e redirects mesmo com client HTTP substituído, manter verificação de IP conectado e aplicar limites bounded de request/response. Testes negativos cobrem redirects e payloads oversized sem depender de uma conta externa.


## 2026-09-22 — OAuth redirect URI allowlist

OAuth start/callback agora exigem allowlist exata por provider, canonicalização, HTTPS e rejeição de fragmentos/userinfo. Loopback HTTP só é permitido com flag explícita e URI allowlisted; o estado PKCE usa a URI canônica.


## 2026-09-22 — session handling web/mobile

O cliente web agentic removeu bearer persistido em localStorage e passou a usar sessão em memória com limpeza 401/403. O mobile limpa SecureStore ao expirar, envia nonce de approval, impede dupla ação e exige confirmação antes de apagar cache/outbox no logout.


## 2026-09-22 — OAuth endpoint egress hardening

OIDC discovery, JWKS, userinfo e token exchange passaram a bloquear proxy/redirects, validar HTTPS e rejeitar IP privado efetivamente conectado. Discovery também valida os endpoints retornados pelo IdP antes de usá-los.


## 2026-09-22 — tool process containment

Terminal e sandbox tools passaram a usar grupo de processos encerrável, limites de timeout/saída, stderr redacted e status explícito de isolamento. O sandbox usa `ulimit` best-effort quando disponível; seccomp/cgroups e sandbox forte permanecem fora do claim.


## 2026-09-22 — plugin/MCP/skill ownership

Connectors, MCP stdio, Remote MCP e skills passaram a carregar ownership opcional por organização; lifecycle autenticado usa métodos scoped e nega cross-tenant/global mutation, enquanto catálogo global permanece read-only.


## 2026-09-22 — CI/release quality gates

O workflow agentic passou a verificar suíte Go completa, vet, build, UI tests/build e mobile typecheck. O release workflow ganhou quality gate obrigatório antes dos builds e da publicação; signing, SBOM e provenance continuam condicionais ao ambiente GitHub/credenciais reais.


## 2026-09-22 — Grok provider contract

Grok passou a validar modelo contra allowlist e catálogo no health probe; a rota HTTP rejeita `stream:true` explicitamente até existir endpoint SSE de produção.


## 2026-09-22 — Compose secure defaults

PostgreSQL/Redis/OTLP foram limitados a loopback por padrão; PostgreSQL/Redis exigem secrets de ambiente e CI usa credenciais efêmeras. O sandbox local não possui Docker, então a integração distribuída permanece dependente do Actions.


## 2026-09-22 — provider selection contract

A Nova tarefa envia e persiste `provider`; apenas `ollama-local` é aceito até adapters Claude/Codex/OmniRoute reais existirem. A UI marca os demais como não conectados.


## 2026-09-22 — release artifact integrity

Release agora falha para artefatos ausentes/vazios e pode gerar attestation de provenance somente com configuração explícita `OLLAMA_ENABLE_ATTESTATIONS=true` no ambiente GitHub.


## 2026-09-22 — tenant isolation de jobs/replay

Fila de jobs agora filtra listagem e replay pela organização da missão, negando cross-tenant sem mutação.


## 2026-09-22 — artifact manifest path safety

Manifests de artefatos agora rejeitam symlinks e resoluções fora do workspace antes de ler e hashear arquivos.
