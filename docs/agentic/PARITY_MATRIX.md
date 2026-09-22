# Matriz de paridade observável — Ollama Classe A+

## Objetivo

Esta matriz define o que o Ollama Classe A+ pretende oferecer como plataforma local-first inspirada em jornadas observáveis de assistentes agentic, incluindo a superfície desktop mostrada no pedido, Claude, Codex, OmniRoute e padrões úteis de outros harnesses. Ela não afirma acesso a internals proprietários nem transforma uma interface, adapter ou mockup em capacidade comprovada.

> **Regra de produto:** uma capacidade só passa a `VALIDADA` quando existe implementação, teste automatizado e execução reproduzível no ambiente correspondente. Credenciais, hardware, contas externas e publicação em lojas são dependências explícitas.

## Estados

| Estado | Significado |
|---|---|
| `VALIDADA LOCALMENTE` | Implementação e testes reproduzíveis no repositório, sem afirmar disponibilidade de serviço externo. |
| `ADAPTER IMPLEMENTADO` | Contrato e adapter existem, mas a execução ponta a ponta depende de provider, credencial, hardware ou infraestrutura externa. |
| `PARCIAL` | Existe uma parte funcional, mas faltam superfícies, estados, integração ou gates essenciais. |
| `PENDENTE` | Ainda não há uma entrega verificável suficiente. |

## Matriz atual

| Domínio | Ollama Classe A+ atual | Meta de paridade observável | Gate que falta |
|---|---|---|---|
| Missões agentic | `VALIDADA LOCALMENTE`: runtime, planner, ferramentas, eventos, recovery, artifacts e approvals em `internal/agent` | Criar, executar, pausar, retomar, corrigir e concluir missões com timeline confiável | Execução distribuída persistente e E2E de jornada longa |
| Shell desktop | `PARCIAL`: UI Agentic Console e rotas base existem | Shell com sidebar, Nova tarefa, Agentes, Habilidades, Plugins, Agendado, Biblioteca, Projetos, Tarefas, Conta e Configurações | Estados reais de cada tela, atalhos, responsividade, acessibilidade e captura E2E |
| Chat local | `VALIDADA LOCALMENTE`: APIs Ollama/OpenAI-compatible e modelos locais | Chat multimodal com streaming, anexos, histórico, projetos e ações agentic | Jornada E2E com dados reais, falhas de modelo e recuperação |
| Claude | `ADAPTER IMPLEMENTADO`: protocolo Anthropic e integração de lançamento/model settings existentes | Selecionar modelos Claude por credencial própria ou gateway, com tools, streaming e limites visíveis | Smoke com credencial real; nunca armazenar segredo em UI, fixture ou Git |
| Codex | `ADAPTER IMPLEMENTADO`: proxy/integração de desktop e catálogo/configuração existentes | Usar Codex como provider ou ferramenta de coding mantendo approval, artifacts e logs | Smoke com sessão/credencial real e validação de limites do provider |
| OmniRoute | `ADAPTER IMPLEMENTADO`: registry OpenAI-compatible, preset local, allowlist de loopback e smoke de forwarding/bearer | Gateway local configurável com `auto`, fallback, quotas, modelos e endpoint `/v1` exibidos na UI | Teste com instância OmniRoute real, health/model discovery e falha de fallback sem violar local-only |
| HarnessRouter | `ADAPTER IMPLEMENTADO`: provider OpenAI Responses-compatible, `harness_id` injetado server-side e preset para Codex/Claude Code | Harnesses plugáveis com sessões, streaming, arquivos, artifacts, cancelamento e seleção por metadata | Instância HarnessRouter real, chave, harness instalado, streaming/follow-up/cancelamento/artifacts e teste de falha |
| Browser Operator | `ADAPTER IMPLEMENTADO`: Playwright, navegação, upload/download, screenshots e takeover | Navegação visual robusta com perfis isolados, approvals e recuperação | Teste com Chromium instalado e jornadas reais em ambiente controlado |
| Computer use / desktop | `ADAPTER IMPLEMENTADO`: companions por capability e pairing | Tela, mouse, teclado, clipboard e processos com grants revogáveis | Testes físicos Linux/macOS/Windows, instaladores assinados e rollback |
| Pesquisa profunda | `VALIDADA LOCALMENTE`: pesquisa multiagente, cache, citações, robots e SSRF guard | Paralelização, síntese com fontes, exportação e retomada | Smoke com fontes externas e orçamento/limites observados |
| Memória/projetos | `VALIDADA LOCALMENTE`: memória episódica/semântica, projetos e ingestão | Retenção, embeddings, recuperação contextual e escopo por organização | Benchmark de recuperação e validação com storage distribuído |
| Skills/MCP | `VALIDADA LOCALMENTE`: manifests, scopes, allowlists e MCP stdio | Instalação, atualização, isolamento, lifecycle e aprovação por ferramenta | Teste com servidores MCP externos e matriz de permissões |
| Desktop Commander Remote MCP | `ADAPTER IMPLEMENTADO`: preset local stdio e Streamable HTTP remoto, HTTPS, allowlist e bearer server-side opcional | Filesystem, terminal, processos e múltiplos dispositivos conectados com approvals | OAuth PKCE, device pairing, conta oficial, agente local e testes físicos; o serviço hospedado está em beta |
| Company OS | `VALIDADA LOCALMENTE`: tenant, identidade, departamentos, roadmap, KPIs, backlog, ciclos, report, budget e pausa por risco | Criar e operar uma empresa com connectors de CRM, social, afiliados, ecommerce, ads e analytics | OAuth/scopes reais, connectors por plataforma, sandbox de anúncios, compliance, pedidos, logística, receita e staging distribuído |
| Automação | `VALIDADA LOCALMENTE`: scheduler, webhooks, retries, DLQ e replay | Agendamentos recorrentes, eventos e tarefas externas com idempotência | Workers persistentes e smoke com serviços externos |
| Integrações | `ADAPTER IMPLEMENTADO`: conectores HTTP allowlisted e contratos OAuth/SSO | GitHub, Google Workspace, e-mail, Slack, Discord, WhatsApp, Jira, Notion e Microsoft 365 | OAuth/refresh real por provider, scopes mínimos e isolamento tenant |
| Multimídia | `ADAPTER IMPLEMENTADO`: imagem, vídeo, speech, transcrição, visão e OCR por provider | Geração/edição local e remota multimodal em uma missão | Modelos instalados, GPU, Tesseract e smoke real por modalidade |
| Builders | `ADAPTER IMPLEMENTADO`: canvas, bindings, eventos, undo/redo, preview e exporters | Drag-and-drop rico para sites, apps, jogos, slides e dashboards | Editor visual completo, colaboração/CRDT, E2E preview/export/deploy |
| Artefatos | `VALIDADA LOCALMENTE`: manifest, hashes, versões e download | Artifacts navegáveis em timeline, diff, preview, export e rollback | E2E com arquivos grandes e permissões por organização |
| Deploy | `ADAPTER IMPLEMENTADO`: Vercel, Netlify e generic com approval e contenção | Publicação real segura com logs, domínio, rollback e health check | Smoke autorizado com contas reais; AWS/Cloudflare ainda requerem adapters específicos |
| Auth/RBAC/SSO | `ADAPTER IMPLEMENTADO`: organizações, RBAC, OAuth/OIDC, SAML, MFA e RLS | Login, recovery, grupos, ABAC/DLP/secrets manager e auditoria enterprise | IdP real, JWKS/discovery, issuer/audience/nonce, rotação e RLS staging |
| Colaboração | `ADAPTER IMPLEMENTADO`: comentários, presença e snapshots | Multiplayer de projetos, missões e builders com conflitos resolvidos | CRDT/realtime E2E e políticas por tenant |
| Mobile | `PARCIAL`: cliente Expo, cache/outbox offline e base de approvals | App Android/iOS com inbox, push, offline sync e device control | Testes físicos, push remoto, resolução de conflitos e distribuição em lojas |
| Observabilidade | `ADAPTER IMPLEMENTADO`: métricas, traces, filas, OTLP e replay | Correlation ID, dashboards, alertas, SLOs e runbooks | Collector/Redis/Postgres em staging e teste de falha/recuperação |

## OmniRoute: decisão de integração

O [OmniRoute](https://github.com/diegosouzapw/OmniRoute) é tratado como **gateway externo opcional**, não como código incorporado ao Ollama Classe A+. A integração correta é por endpoint compatível e credencial server-side, por exemplo um endpoint local configurado pelo operador. O Classe A+ deve:

1. permitir declarar o endpoint e o modelo sem expor token;
2. limitar HTTP não seguro a loopback explicitamente permitido, mantendo HTTPS obrigatório para hosts externos;
3. consultar modelos/health somente quando o operador habilitar a integração;
4. preservar `local-only` para nunca desviar uma missão marcada como local;
5. registrar provider, modelo, latência, falha e fallback sem registrar prompt ou segredo indevido;
6. exigir approval para mudanças de routing, conexão de contas e ações externas.

Um provider OpenAI-compatible existente não prova que a integração OmniRoute foi validada. O gate será um teste contra uma instância local real, com modelo de teste, auth configurada, falha simulada e verificação de que o fallback não viola a política local-only.

## HarnessRouter: decisão de integração

O [HarnessRouter Community Edition](https://github.com/HarnessRouter/harnessrouter) é um complemento particularmente útil para a meta de reunir vários coding harnesses. Ele implementa uma interface unificada e compatível com OpenAI Responses para harnesses plugáveis, com sessões, streaming, arquivos, artifacts, cancelamento e falhas estruturadas. A integração do Classe A+ está detalhada em [`HARNESSROUTER.md`](HARNESSROUTER.md) e no preset [`examples/dz23-harnessrouter.json`](../../examples/dz23-harnessrouter.json).

O Classe A+ não incorpora o código do HarnessRouter nem trata a licença Apache-2.0 como licença dos CLIs, modelos ou Starter Kits executados por ele. A conexão é opt-in, server-side e substituível. O `harness_id` é configurado por modelo e injetado no campo `metadata` pelo proxy, enquanto chaves ficam fora do browser. O gate de promoção para `VALIDADA LOCALMENTE` exige instância real, provider e harness instalados, além de streaming, follow-up de sessão, cancelamento, artifact e falha/recovery reproduzíveis.

## Critério de conclusão da paridade

A meta não é declarar “igual ao Manus” por nomes de menus. A meta é completar jornadas observáveis equivalentes e adicionais:

- criar uma missão a partir de Nova tarefa;
- escolher modelo local, Claude, Codex ou OmniRoute;
- pesquisar, usar browser, executar código e pedir approval quando necessário;
- produzir artifacts, diff, preview, export e, se aprovado, deploy;
- acompanhar timeline, custos/quotas, traces, erros e recuperação;
- retomar no desktop e no mobile com memória e permissões corretas;
- repetir a jornada sem cross-tenant, SSRF, path traversal, vazamento de segredo ou bypass de approval.

A matriz deve ser atualizada a cada rodada junto com testes, screenshots e `CHANGELOG.md`. Uma linha só pode avançar de estado com evidência reproduzível.
