# Roadmap executável — Ollama DZ23 Agentic Platform

## Fase 0 — Base, contratos e gates — CONCLUÍDA LOCALMENTE

Consolidar os contratos versionados, checkpoint, build, testes, threat model, política de secrets e observabilidade. Esta fase inclui a ponte Claude/Codex já implementada e o primeiro pacote de documentação agentic.

## Fase 1 — Núcleo vertical de missões — CONCLUÍDA LOCALMENTE

Implementar o store, estados, eventos, planner, executor, observer, recovery, aprovação e artifact manifest. O fluxo comprovado será: criar missão, gerar plano, executar leitura autorizada, persistir resultado, produzir manifesto e consultar eventos.

## Fase 2 — Ferramentas seguras — CONCLUÍDA LOCALMENTE

Adicionar filesystem por workspace, terminal com executável e argumentos separados, execução de código em sandbox, limites de CPU/memória/tempo/rede, logs redacted e kill seguro. A execução arbitrária permanece bloqueada até a policy fornecer scope e aprovação.

## Fase 3 — Projetos, memória, skills e MCP — CONCLUÍDA LOCALMENTE

Criar projetos com permissões e contexto, memória episódica e semântica com retenção configurável, importação de arquivos, skill manifests assinados ou confiáveis e lifecycle de MCP servers. Cada fonte de contexto terá origem e nível de confiança.

## Fase 4 — Jobs e automações — CONCLUÍDA LOCALMENTE

Adicionar fila persistente, scheduler, retries, idempotency keys, webhooks verificados, dead-letter queue e replay. Integrações externas serão adapters com secrets server-side, scopes mínimos e confirmação para efeitos sensíveis.

## Fase 5 — Browser e computer use — ADAPTERS IMPLEMENTADOS; PROVA NATIVA PENDENTE

Integrar browser isolado com perfis, downloads, uploads, navegação, screenshots, ações e takeover humano. Integrar Desktop companion para tela, mouse, teclado, clipboard e processos usando capability grants. Nenhuma dessas capacidades será simulada por texto.

## Fase 6 — Artefatos multimídia e builders — ADAPTERS IMPLEMENTADOS; PROVIDERS/DEPLOY PENDENTES

Adicionar documentos, slides, planilhas, gráficos, dashboards, imagens, áudio, voz, transcrição, vídeo, sites, aplicativos e jogos. Cada domínio deve possuir renderer, preview, export, manifest, checksum e smoke test.

## Fase 7 — Superfícies de produto — PARCIAL

Construir API pública versionada, CLI agentic, Web/Desktop com timeline de missão, diff, approvals, logs, artifacts e terminal controlado. Construir aplicativo Mobile para chat, inbox de missões, approvals, notifications, artifacts e controle de projetos.

## Fase 8 — Integrações e colaboração — PARCIAL

Adicionar GitHub, Google Workspace, e-mail, Slack, Discord, WhatsApp e outros connectors por adapters. Implementar organizações, usuários, papéis, RBAC/ABAC, auditoria, compartilhamento, comentários, presença e colaboração em tempo real.

## Gates por fase

Cada fase exige testes unitários, contratos, autorização negativa, integração real quando o adapter existir, segurança, observabilidade e documentação. Uma feature fica `PARTIAL` enquanto seu adapter não tiver execução real e smoke reproduzível.

## Ordem de investimento

A prioridade é segurança e recuperação, depois execução real, depois conectores e superfícies. A interface não será usada para mascarar lacunas de runtime. O próximo incremento executável é substituir os adapters locais por stores/filas distribuídos, configurar providers OAuth/media/deploy, executar smoke tests em Windows/macOS/Android/iOS e fechar o fluxo de publicação no GitHub.


## Incremento 2026-09-21 — Multiagente, pesquisa e dispositivos

Foi implementado um AgentOrchestrator com sete papéis especializados, concorrência limitada, retries, orçamento, cancelamento, persistência e reducer com conflitos. O ResearchEngine agora executa pesquisa multi-fonte com cache, citações, hash, extração HTML, robots policy e SSRF guard. O DeviceStore adiciona pairing one-time, capability report, heartbeat, listagem e revogação. O próximo incremento deve conectar o reducer a um modelo de síntese validado, adicionar fontes PDF/OCR, WebSocket/mTLS e testes físicos dos companions.


## Incremento 2026-09-21 — Infraestrutura distribuída e transporte seguro

A fase adicionou adapters opcionais de PostgreSQL, Redis e OpenTelemetry, WebSocket de companion com TLS/mTLS policy, stack Docker Compose de desenvolvimento, workflow de qualidade/SBOM e fila offline no mobile. A próxima etapa de produção deve validar Redis/PostgreSQL reais em CI, configurar RLS/tenant isolation, rotação mTLS, push remoto, resolução de conflitos mobile, exporters persistentes e assinatura/rollback das releases.


## Incremento 2026-09-21 — SSO enterprise, credenciais por tenant e histórico visual

A fase adicionou adapter SAML baseado em `crewjam/saml`, incluindo metadata, AuthnRequest assinado, RelayState one-time, ACS e provisionamento de claims; conectores com resolução de access token OAuth cifrado por organização; RLS PostgreSQL forçado para impedir bypass pelo dono da tabela; TLS 1.3/mTLS opcional com recarga de certificado por handshake; e canvas visual com bindings, eventos, histórico persistente, undo e redo. Permanecem pendentes os testes ponta a ponta com IdP, serviços distribuídos reais fora do CI, providers multimídia/deploy, testes físicos de companions e distribuição assinada.


## Incremento 2026-09-21 — Publicação real de builders

Foi adicionado o DeploymentManager com adapters Vercel, Netlify e generic, coleta segura de arquivos, limites, redirects bloqueados, token server-side e approval explícito. O próximo gate é executar smoke contra contas reais e completar adapters de AWS/Cloudflare conforme credenciais e requisitos de cada ambiente; a publicação externa nunca é simulada como concluída apenas por existir um preview local.

## Incremento 2026-09-22 — Árvore de produto, shell desktop e OmniRoute

Foi criada a árvore pública [`PRODUCT_TREE.md`](PRODUCT_TREE.md), que separa a superfície observável de um desktop agentic, a árvore atual do Classe A+ e a árvore-alvo unificada com Claude, Codex, OmniRoute, coding agents, pesquisa, builders, mobile, colaboração e operações. A matriz [`PARITY_MATRIX.md`](PARITY_MATRIX.md) passa a ser o contrato de honestidade para diferenciar `VALIDADA LOCALMENTE`, `ADAPTER IMPLEMENTADO`, `PARCIAL` e `PENDENTE`.

O shell web ganhou rotas reais para Biblioteca, Projetos, Agendado, Habilidades, Plugins e Tarefas, além de uma navegação lateral com Nova tarefa, Agente, workspace e conta. A Settings ganhou o Agentic Control Center com catálogo de modelos, status sanitizado de runtime, approvals, isolamento, conectores, MCP, mídia, deploy e OTLP. O endpoint `/api/agent/v1/config/safe` nunca retorna tokens, caminhos privados ou valores de configuração.

O multi-provider agora aceita HTTP somente para um serviço numérico de loopback quando `allow_private` e `allow_insecure_loopback` estão explicitamente definidos. Isso permite o preset [`examples/dz23-omniroute.json`](../../examples/dz23-omniroute.json) para uma instância OmniRoute local, mantendo HTTPS obrigatório para hosts externos. Foram adicionados testes de carregamento, rejeição de HTTP externo e encaminhamento com bearer server-side. A integração continua dependente de uma instância OmniRoute, credencial e smoke test do operador.

As telas novas têm estados vazios honestos e ações de entrada. CRUD persistido, colaboração, editor drag-and-drop, provider discovery completo e deploy real continuam como próximos gates; a interface não os apresenta como concluídos.


## Incremento 2026-09-22 — fluxo vertical, home e HarnessRouter

A rodada substituiu estados estáticos do shell por contratos reais: Projetos e Agendado possuem criação, listagem, edição/exclusão com escopo de organização; Tarefas e Biblioteca consultam missões/artifacts; Plugins e Habilidades consultam catalogs server-side; e o Agentic Console cria missões com provider (Ollama local, Claude, Codex, OmniRoute) e projeto selecionáveis. A home agora possui composer, recomendações e atalhos funcionais, com fallback offline local-first.

O proxy multi-provider ganhou `harness_id` declarativo por modelo e injeção server-side de `metadata.harness_id`, permitindo o preset opcional [`examples/dz23-harnessrouter.json`](../../examples/dz23-harnessrouter.json) para Codex e Claude Code no [HarnessRouter](https://github.com/HarnessRouter/harnessrouter). Isso é adapter implementado; não é validação de uma instância ou credencial externa.

Provas desta rodada: `go test ./internal/agent ./server`, build Vite, smoke CRUD real contra servidor local, smoke de missão/eventos, captura Chromium das nove rotas e `node app/ui/app/scripts/smoke-shell.mjs` com home, Projects, Scheduled, Plugins, Skills e seleção Claude. O próximo gate é HarnessRouter real com streaming/follow-up/cancelamento/artifacts, seguido de editor drag-and-drop/CRDT e testes distribuídos/IdP/dispositivos.


## Incremento 2026-09-22 — Company OS e Desktop Commander Remote MCP

A base passou a incluir um Company OS persistente. Uma empresa é criada no tenant ativo e recebe identidade, posicionamento, modelo de negócio, departamentos virtuais, roadmap, metas, backlog, ciclos, budget, relatório e risco. Ciclos são ligados a schedules persistentes e o runtime interrompe a criação de novas missões quando a empresa está pausada por ação do operador, orçamento excedido ou anomalia grave.

Também foi implementado um adapter Remote MCP Streamable HTTP e o preset oficial do Desktop Commander. O runtime valida HTTPS fora de loopback, mantém allowlist de métodos, usa bearer apenas por variável server-side e expõe `mcp.remote.call` com approval. O transporte não é confundido com autenticação: OAuth PKCE, conta, device pairing, agente `npx ... remote`, revogação e testes físicos permanecem dependentes do operador e do serviço oficial.

Os próximos gates do Company OS são connectors reais com CRM, e-mail, redes sociais, anúncios, afiliados, ecommerce, logística e analytics, sempre com sandbox, scopes mínimos, DLP, approval e testes de compliance. Os próximos gates do Remote MCP são completar um fluxo OAuth autorizado, parear um dispositivo de teste, executar apenas jornadas reversíveis e validar revogação, logs e desligamento. Nenhum desses gates é simulado pela presente revisão.
