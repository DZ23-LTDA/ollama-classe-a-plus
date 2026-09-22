# Relatório de prontidão — Ollama Classe A+

**Data:** 2026-09-22
**Repositório:** `DZ23-LTDA/ollama-classe-a-plus`
**Branch auditada:** `feat/manus-parity-omniroute`
**Base local:** commit `dfe07d5665b11f025fb36c6382298e8c053733f5`

## Veredito executivo

O projeto está **pronto como release candidate local-first para instalação, desenvolvimento e validação controlada**. Ele não está pronto para ser classificado como produto final de produção universal, porque ainda faltam validações com contas externas, infraestrutura distribuída real, dispositivos físicos, modelos multimídia locais, assinatura de instaladores, lojas e operações comerciais.

A distinção é importante: o runtime, a UI, os contratos, os guardrails e as jornadas sandbox podem ser executados e testados. Já uma integração com uma conta real de xAI, Composio, TikTok Shop, Meta, Shopify ou Desktop Commander não pode ser considerada concluída apenas porque existe um adapter ou uma tela.

## Teste das APIs fornecidas

As chaves do arquivo fornecido foram usadas somente em chamadas curtas de leitura/diagnóstico e em uma inferência gratuita do OpenRouter. Nenhuma chave foi copiada para o repositório, para logs públicos, screenshots ou este relatório. O arquivo fornecido continha credenciais em texto aberto; todas devem ser consideradas expostas e rotacionadas pelo operador.

| Serviço | Verificação realizada | Resultado | Interpretação |
|---|---|---:|---|
| OpenRouter | `GET /api/v1/models` e inferência gratuita | HTTP 200 | Chave respondeu; gateway Classe A+ também encaminhou uma chamada gratuita ponta a ponta |
| Groq | `GET /openai/v1/models` | HTTP 200 | Chave respondeu ao catálogo |
| DeepSeek | `GET /models` | HTTP 200 | Chave respondeu ao catálogo |
| Fireworks | `GET /inference/v1/models` | HTTP 200 | Chave respondeu ao catálogo |
| Cerebras | `GET /v1/models` | HTTP 200 | Chave respondeu ao catálogo |
| Mistral | `GET /v1/models` | HTTP 200 | Chave respondeu ao catálogo |
| NVIDIA | `GET /v1/models` | HTTP 200 | Chave respondeu ao catálogo |
| Novita | `GET /openai/v1/models` | HTTP 200 | Chave respondeu ao catálogo |
| Cohere | `GET /v1/models` | HTTP 200 | Chave respondeu ao catálogo |
| Google Gemini | `GET /v1beta/models` com `x-goog-api-key` | HTTP 200 | Chave respondeu ao catálogo |
| Hugging Face | `GET /api/whoami-v2` | HTTP 200 | Token respondeu à identificação |
| GitHub | `GET /user` | HTTP 200 | Token respondeu à identificação; revisar escopos e rotacionar |
| Cloudflare | `GET /user/tokens/verify` | HTTP 200 | Token respondeu à verificação |
| Together | `GET /v1/models` | HTTP 401 | Chave inválida, expirada ou sem permissão para esse endpoint |
| xAI | `GET /v1/models` | HTTP 403 | A chave foi recusada; não classificar xAI como conectado |
| Hyperbolic | `GET /v1/models` | HTTP 401 | Chave inválida, expirada ou sem permissão |
| Alibaba/DashScope | `GET /compatible-mode/v1/models` | HTTP 401 | Chave inválida, expirada ou endpoint/escopo incompatível |
| Voyage | endpoint de catálogo tentado | HTTP 404 | Endpoint escolhido não prova validade ou invalidade da chave; requer health check específico |

O smoke escolheu o modelo gratuito `inclusionai/ling-3.0-flash-vl:free` no OpenRouter. A chamada direta retornou HTTP 200. Em seguida, o servidor Ollama Classe A+ foi iniciado em uma porta isolada, publicou o modelo remoto em `/v1/models` e devolveu HTTP 200 através de `/v1/chat/completions`. Esse é o principal teste externo ponta a ponta desta rodada.

## Gates locais

Os gates locais abaixo passaram antes ou durante esta rodada: `scripts/check-class-a-plus-integrity.sh`, `CGO_ENABLED=1 go test ./... -count=1`, `CGO_ENABLED=1 go vet ./...`, build Go, Vitest da UI com 20 arquivos e 199 testes, build Vite, typecheck Expo, smoke Chromium do shell, smoke Growth OS e smoke Builder com approval. O teste específico `TestBrowserOperatorNavigateAndSnapshot` também passou localmente após a correção do CI.

O gate distribuído com PostgreSQL, Redis e OTLP foi **N/A neste sandbox**, porque Docker não está instalado. O workflow público possui job para essa validação; a conclusão deve ser verificada no GitHub após a execução do novo commit.

## Correção interna desta rodada

O workflow `.github/workflows/dz23-agentic-quality.yaml` agora instala Playwright e Chromium antes de executar os testes do Browser Operator e exporta o executável descoberto para o processo Go. O adaptador `internal/agent/browser.go` também passou a incluir stderr sanitizado no erro, tornando falhas de dependência diagnosticáveis sem imprimir segredos.

## O que falta para o produto final de produção

| Prioridade | Pendência | Por que ainda não está pronta |
|---|---|---|
| P0 | Rotação das chaves fornecidas | Foram entregues em texto aberto; devem ser revogadas e recriadas antes de qualquer uso contínuo |
| P0 | CI remoto verde no novo commit | O commit anterior falhou no Browser Operator por dependência ausente; a correção precisa ser executada no GitHub |
| P0 | Testes de autorização e isolamento em staging distribuído | O compose, RLS, Redis e OTLP existem, mas precisam de PostgreSQL/Redis/Collector reais e execução repetível |
| P1 | Composio e Desktop Commander | Adapter/preset existem; OAuth, sessão, pairing, dispositivo e uma tool real ainda precisam de conta de teste autorizada |
| P1 | xAI/Grok | O adapter existe; a chave fornecida recebeu 403, portanto Responses, tools e quota xAI não foram validados |
| P1 | Social commerce | Growth OS é sandbox; TikTok Shop, Meta, Instagram, Shopify, YouTube e WhatsApp exigem apps, escopos, webhooks, região e aprovação |
| P1 | Desktop e mobile físicos | Linux/macOS/Windows e Android/iOS precisam de execução em dispositivos/runners reais, push, offline e rollback |
| P1 | Mídia local | OCR, visão, imagem, vídeo, TTS e STT precisam de modelos instalados, hardware e testes de qualidade |
| P1 | Builder e deploy externo | Preview/export/deploy local estão testados; Vercel, Netlify, AWS, Cloudflare e DNS exigem contas e smoke reversível |
| P2 | Release de distribuição | Faltam instaladores assinados, SBOM/provenance verificável, auto-update/rollback e publicação nas lojas |
| P2 | Hardening adicional | Ainda são desejáveis `openat`/handles contra TOCTOU, atestados assinados de skills, egress enforcement completo e storage seguro de credenciais do renderer |

## Conclusão operacional

O projeto pode ser entregue agora para **desenvolvimento local, avaliação do runtime, testes de providers com credenciais rotacionadas e construção de jornadas agentic**. Ele não deve ser vendido ou documentado como uma plataforma já validada em produção, como uma cópia interna do Manus/Grok/Claude/Codex, nem como uma operação comercial autônoma sem os gates externos acima.

A próxima barreira objetiva é publicar a correção do Browser Operator, aguardar o CI e, depois, executar staging distribuído e integrações externas uma por uma, sempre com escopos mínimos, approval, idempotência, rollback e nenhuma chave no Git.


## Hardening posterior ao smoke de APIs — 2026-09-22

A rodada posterior fechou quatro riscos internos que ainda eram ajustáveis sem credenciais externas. Chamadas de connector agora exigem `organization_id` no caminho tenant-aware; o matching de `path_prefixes` respeita limites de segmento; e o transporte padrão rejeita destinos privados resolvidos por DNS, mantendo loopback somente quando o próprio `BaseURL` é loopback. O Remote MCP passou a validar nomes de ambiente e bloquear headers de transporte que poderiam interferir no framing HTTP. O MCP stdio aplica a mesma validação às variáveis herdadas.

Alterações globais de lifecycle de connectors, MCP, Remote MCP e skills agora exigem papel `owner` ou `admin` quando o runtime está autenticado. O modo local-first sem autenticação continua disponível somente no bind loopback já protegido pelo middleware. O Browser Operator passou a escolher um fallback conhecido de Chromium quando um caminho configurado não existe; o teste específico passou localmente inclusive com caminho configurado inválido.

Os gates desta iteração passaram localmente: integrity guard, `CGO_ENABLED=1 go test ./... -count=1`, `CGO_ENABLED=1 go vet ./...`, build Go, testes UI (20 arquivos/199 testes), build UI, typecheck Expo mobile e teste Browser Operator isolado. A nova revisão pública ainda precisa concluir no GitHub; o resultado remoto deve ser associado ao SHA desta iteração, não ao workflow histórico.

Essas correções não convertem adapters em integrações externas conectadas. Continuam pendentes, por dependerem de autoridade, contas, ambientes ou dispositivos reais, os testes distribuídos PostgreSQL/RLS, Redis e OTLP, OAuth/SSO contra IdP, Composio, xAI, Desktop Commander Remote, canais sociais e marketplaces, GPU/modelos multimídia, runners físicos, assinatura/lojas e deploy externo.


## Slice P0 de ownership e invariantes Company — 2026-09-22

O primeiro slice P0 após a auditoria ampliada fechou uma fronteira específica, não a auditoria inteira. Projetos Builder agora carregam `organization_id`; listagem, lookup, mutação visual, undo/redo, preview, export, publish, deploy e preview de arquivo passam por lookup scoped no servidor. O entry precisa existir, nomes usados em templates HTML são escapados e preview/export recusam symlinks que resolvam fora do projeto. O Company OS passou a criar empresas por DTO allowlisted, resetando ID, status, coleções, gasto, risco e timestamps server-managed; gasto de agente pausado ou acima do budget falha sem incrementar o valor gasto.

A evidência reproduzível desta slice é o teste HTTP `server/builder_scope_test.go` com duas organizações, que confirma listagem vazia e `403` sem mutação para o tenant errado; testes de Builder para entry ausente, nome com XSS, symlink e organização; e testes Company para create privilegiado e gasto atômico. Os gates completos passaram: integrity guard, `CGO_ENABLED=1 go test ./... -count=1`, `go vet`, build Go, Vitest (20 arquivos/199 testes), build UI e typecheck mobile.

Ainda não é produção-ready. Permanecem P0 em orchestration/traces/devices/pairing e demais objetos, approvals com separação forte de aprovador, sandbox/MCP process isolation, SSRF/DNS rebinding e DLP, lifecycle pausado de todas as mutações Growth/Social, bearer/redirect/session handling, CI/release e validações distribuídas/externas. A classificação continua **preview/local RC em hardening**.


## Slice P0 de orchestration, traces e devices — 2026-09-22

A segunda slice pós-auditoria adicionou `organization_id` aos orchestration jobs e propagou o tenant aos spans de missão e ferramenta. As rotas de criar, ler, executar e cancelar orchestration, bem como listagem global de traces, agora usam o tenant da requisição. Devices e pairing passaram a filtrar listagem, heartbeat e revoke por organização; o código de pairing fica vinculado à organização que o criou e não pode ser sobrescrito por outro tenant. Regressões negativas confirmam `403` e ausência de mutação para reads, cancelamento, revoke e heartbeat cross-tenant.

A evidência inclui `server/p0_scope_test.go`, testes de Orchestrator, TraceStore e DeviceStore, além dos gates completos: integrity guard, `CGO_ENABLED=1 go test ./... -count=1`, `go vet`, build Go, Vitest 20/199, build UI e typecheck mobile. O handshake WebSocket continua autenticado pelo device token; mTLS/TLS, pairing físico e testes em companions reais permanecem dependências externas.

Esta slice não encerra a auditoria. Continuam P0 em plugins/MCP/skills/artifacts, approvals com separação forte de aprovador, sandbox/process isolation, SSRF/DNS rebinding e DLP, lifecycle pausado de todas as mutações Growth/Social, sessão segura no renderer e release supply chain. A classificação permanece **preview/local RC em hardening**.


## Slice P0 de approvals de missão — 2026-09-22

A terceira slice pós-auditoria endureceu approvals de missão. Quando autenticação está ativa, somente `owner` ou `admin` da organização pode decidir; a decisão passa por `DecideApprovalForActorCAS`, exige a versão corrente da missão e consome o nonce emitido para aquela aprovação. Replays, nonce incorreto, organização errada, razão vazia e decisão já consumida são rejeitados. O Agentic Console passou a enviar o nonce retornado pela API.

A evidência inclui teste de CAS/nonce no Runtime, teste de role policy no servidor e os gates completos: integrity guard, `CGO_ENABLED=1 go test ./... -count=1`, `go vet`, build Go, Vitest 20/199, build UI e typecheck mobile. O modo local sem autenticação continua disponível para instalação loopback e não deve ser confundido com autorização multiusuário.

Esta slice cobre approvals de missão, não todas as decisões do Company/Growth/Social OS. Endpoints que ainda aceitam campos booleanos de aprovação para campanhas, pedidos, drafts ou gasto permanecem pendência P0/P1 e não devem ser tratados como autoridade auditável para efeitos externos.


## Slice P0 de approvals auditáveis do Company OS — 2026-09-22

Campanhas, programas de afiliados, pedidos de dropshipping e drafts sociais agora criam uma decisão server-side `CompanyApproval` com organização, recurso, policy, nonce, expiração, actor, razão e status. Os endpoints de approve localizam o approval pendente no servidor, exigem owner/admin quando auth está ativa, validam nonce e versão da Company e só então projetam `approved=true`/estado do recurso. Regressões cobrem nonce incorreto, replay, actor/org e HTTP.

Os gates completos passaram: integrity guard, `CGO_ENABLED=1 go test ./... -count=1`, `go vet`, build Go, Vitest 20/199, build UI e typecheck mobile. O campo booleano `approved` continua no schema apenas como projeção legada; o cliente não é mais a autoridade para aprovar. `RecordSpend`/gasto da Company ainda aceita uma rota distinta com booleano e permanece P0 aberto para migrar ao mesmo approval ledger.


## Slice P0 final de approvals de gasto — 2026-09-22

A rota de gasto foi migrada para `RecordSpendRequest`: quando a política exige aprovação, a solicitação cria um approval `spend` pendente e retorna `202 Accepted` sem debitar o budget. A fila `CompanyApprovalQueue` permite owner/admin decidir com nonce e CAS por endpoint genérico; somente após decisão aprovada o valor é contabilizado. Rejeição, replay, organização incorreta, versão conflitante e excesso de budget não produzem débito parcial.

Os gates completos após a correção de contrato TypeScript passaram: integrity guard, `CGO_ENABLED=1 go test ./... -count=1`, `go vet`, build Go, Vitest 20/199, build UI e typecheck mobile. O método interno legado `RecordSpend(..., approved bool)` permanece apenas para compatibilidade de domínio/testes; a rota HTTP e a UI não o utilizam mais nem aceitam essa autoridade do cliente.
