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


## Slice P0 de DLP em resultados e observabilidade — 2026-09-22

`RedactValue` passou a redigir recursivamente strings, mapas e listas, além de valores sob chaves sensíveis como `token`, `secret`, `password`, `api_key` e `private_key`. Foram adicionados padrões para PEM, GitHub, OpenAI, OpenRouter, xAI, AWS, Slack, Bearer e assignments de credenciais. O runtime aplica a redação a `Step.Result`, erros de step/mission e payloads de eventos; TraceStore protege atributos, nomes e erros; JSONStore e PostgresStore protegem serializações persistidas.

Testes negativos injectam tokens em resultado de tool, evento, trace e missão persistida e verificam que nenhum token cru reaparece. Os gates completos passaram: integrity guard, Go tests/vet/build, Vitest 20/199, UI build e mobile typecheck. Payloads de saída para connectors/MCP ainda exigem uma política separada para distinguir segredo operacional autorizado de dado sensível do usuário; essa lacuna não foi declarada resolvida.


## Slice P0 de egress Remote MCP — 2026-09-22

O transporte Remote MCP agora desabilita proxy ambiental, mantém redirects somente no mesmo origin permitido e verifica o endereço IP efetivamente conectado pelo socket TCP. Um hostname que resolver ou rebinding para loopback, privado, link-local, multicast ou unspecified é rejeitado no dialer externo; loopback explícito continua disponível apenas para servidores locais. Testes cobrem redirect same-origin e conexão privada real, além dos allowlists existentes.

Os gates completos passaram: integrity guard, Go tests/vet/build, Vitest 20/199, UI build e mobile typecheck. A prova é local; ainda são necessários testes de rede distribuída, TLS/certificados, DNS controlado e auditoria do egress de Media/Connectors para classificar a plataforma como production-ready.


## Slice P0 de MCP stdio lifecycle — 2026-09-22

MCP stdio agora exige caminho absoluto para um executável regular não-symlink, limita quantidade/tamanho de argumentos, usa ambiente mínimo com apenas variáveis explicitamente allowlisted e sempre inicia em diretório de trabalho explícito privado. Quando o diretório não é informado, um workspace temporário é criado e removido no stop; após timeout/cancelamento ele é recriado para permitir restart limpo. Requests e responses têm limite de 4 MiB, stderr fica limitado a 64 KiB e é redigido antes de aparecer em erros.

Em Unix, o processo é iniciado em grupo próprio e o stop tenta encerrar apenas esse grupo, com fallback seguro ao processo individual; Windows usa o kill nativo. Testes cobrem comando relativo, symlink, workspace cleanup, payload limit, cancelamento e restart. Os gates completos passaram: integrity guard, Go tests/vet/build, Vitest 20/199, UI build e mobile typecheck.

A classificação permanece **contenção best-effort**, não sandbox forte. Ainda faltam seccomp, cgroups, rlimits, limites de PID/memória/CPU e validação física multi-plataforma para afirmar isolamento forte.


## Slice P0 de Media egress/download — 2026-09-22

O MediaManager passou a usar transport sem proxy ambiental, dialer com verificação do IP efetivamente conectado e redirects desabilitados. Downloads de provider aceitam HTTPS ou HTTP explícito em loopback, rejeitam userinfo/fragments, fazem leitura bounded com detecção de overflow e validam `Content-Type` contra a extensão e magic bytes de PNG, JPEG, WebP, MP4 e WAV antes de materializar o artefato. Respostas JSON e áudio também têm limites explícitos.

Regressões locais cobrem redirect, MIME mismatch, magic inválido, payload acima do limite e socket privado. Os gates completos passaram: integrity guard, Go tests/vet/build, Vitest 20/199, UI build e mobile typecheck. O upload de transcrição ainda é limitado por tamanho de entrada, e testes de rede distribuída/TLS real continuam necessários.


## Slice P0 de Connector egress parity — 2026-09-22

Connectors agora aplicam transport sem proxy ambiental também quando o client é substituído por um transport HTTP customizado, preservam o dialer que verifica o IP efetivamente conectado, bloqueiam redirects em cada request e limitam request body a 1 MiB e response body a 2 MiB com detecção de overflow. Regressões cobrem redirect, resposta oversized e request oversized, além dos testes de OAuth tenant-scoped existentes.

Os gates completos passaram: integrity guard, Go tests/vet/build, Vitest 20/199, UI build e mobile typecheck. O contrato de egress não redige credenciais operacionais autorizadas; ainda é necessário separar classificação de dados de usuário, injeção de segredo e payload externo, além de validar uploads e redes distribuídas reais.


## Slice P0 de OAuth redirect URI hardening — 2026-09-22

OAuth providers agora carregam `OLLAMA_AGENT_OAUTH_<PROVIDER>_REDIRECT_URIS` como allowlist separada por vírgula, ponto e vírgula ou linha, e só aceitam no start/callback uma URI canônica exatamente presente nessa lista. Redirects exigem HTTPS; HTTP só pode ser usado para loopback quando `..._ALLOW_LOOPBACK_REDIRECT=true` e a URI loopback também está na allowlist. Userinfo, fragmentos e URIs opacas são rejeitados. O AuthStore aplica sintaxe segura e grava a forma canônica no estado PKCE, reduzindo mismatch e evitando confiança em um valor arbitrário do request.

Testes cobrem host não allowlisted, HTTP externo, fragmento, userinfo, loopback explícito e estado inseguro. Os gates completos passaram: integrity guard, Go tests/vet/build, Vitest 20/199, UI build e mobile typecheck. A validação de issuer/audience/nonce OIDC já existente permanece separada; endpoints OAuth ainda precisam ser alinhados ao mesmo egress/IP/redirect policy em uma slice posterior.


## Slice P0 de session handling web/mobile — 2026-09-22

O cliente web agentic deixou de ler bearer e organização de `localStorage`: a sessão é mantida somente em memória por `setAgentSession`, pode ser removida por `clearAgentSession` e limpa automaticamente em respostas 401/403, emitindo evento local para a UI reagir. Testes verificam que o header é formado sem acessar localStorage, que token vazio é rejeitado e que sessão expirada é apagada.

No mobile, 401/403 removem o token do Expo SecureStore sem enfileirar mutações não autorizadas; approvals enviam nonce, ficam desabilitados enquanto uma decisão está em andamento e receberam labels de acessibilidade. Logout exige confirmação quando há missão/cache/outbox e, após confirmação, remove sessão, missão, eventos, push marker e ações offline. A validação de armazenamento seguro nativo depende de build/dispositivo real; a UI web ainda requer uma jornada de login que chame `setAgentSession` em vez de persistir credenciais.


## Slice P0 de OAuth endpoint egress — 2026-09-22

Discovery OIDC, JWKS, userinfo e token exchange passaram a usar um client seguro que remove proxy ambiental, bloqueia redirects e verifica o IP efetivamente conectado, aceitando loopback apenas quando o endpoint configurado é explicitamente loopback. URLs de endpoints rejeitam credentials/fragments e exigem HTTPS; discovery valida issuer e também os endpoints de autorização, token, JWKS e userinfo retornados antes de usá-los.

Regressões cobrem redirect OAuth e socket privado real, além de validação de endpoint existente. Os gates completos passaram: integrity guard, Go tests/vet/build, Vitest 20/199, UI build e mobile typecheck. Um IdP real/staging ainda é necessário para validar discovery, assinatura, audience, nonce, rotação e refresh ponta a ponta.


## Slice P0 de contenção best-effort de tools — 2026-09-22

`terminal.exec` e `sandbox.exec` agora iniciam processo em grupo próprio e encerram o grupo ao cancelar ou atingir timeout, evitando deixar filhos órfãos no caminho comum. Saídas stdout/stderr são limitadas e passam por `RedactDLP` antes de retornar; o resultado informa a postura `best-effort-process-group` ou `best-effort-unshare`. O sandbox `unshare` aplica, quando suportado pelo shell/kernel, limites best-effort de CPU, memória virtual, processos, descriptors e tamanho de arquivo via `ulimit`, além de continuar com user/mount/PID/network namespaces e ambiente mínimo.

Testes cobrem redaction de stderr, status explícito, cancelamento rápido e encerramento do grupo; a suíte `internal/agent` e os gates completos passaram. Isso não é sandbox forte: seccomp, cgroups, enforcement robusto de PID/memória/CPU e validação real em Windows/macOS/Linux continuam pendentes. O produto deve mostrar essa limitação ao operador em vez de alegar isolamento equivalente a E2B.


## Slice P0 de ownership de plugins/MCP/skills — 2026-09-22

Connector, MCP stdio, Remote MCP e Skill manifests agora possuem `organization_id` quando são tenant-owned. Catálogos autenticados mostram recursos tenant-owned da organização ativa e preservam recursos sem owner como configuração global read-only; operações enable/disable/remove autenticadas exigem ownership exato e retornam `403` para outro tenant ou para recurso global, sem mutação. O modo local sem autenticação continua usando lifecycle compatível no bind loopback.

Regressões criam recursos em `org-a` e `org-b`, confirmam filtragem de lista e rejeitam mutation cross-tenant para Connector, MCP, Remote MCP e Skill. Os gates completos passaram: integrity guard, Go tests/vet/build, UI Vitest/build e mobile typecheck. Ainda faltam endpoints server-owned para registrar novos recursos por organização, assinatura/attestation de skills, isolamento de execução MCP por tenant e prova distribuída com Postgres/RLS.


## Slice P0 de CI/release quality gates — 2026-09-22

O workflow `dz23-agentic-quality` foi corrigido para YAML válido, removeu `npm ci` sem lockfile do mobile, e passou a executar typecheck, Vitest e build de produção da UI além de typecheck mobile, integrity, suíte Go completa, vet e build. O Browser Operator mantém instalação e fallback explícito do Chromium. O workflow de release ganhou um job `quality` independente com esses gates e todos os builds/publicação dependem dele; isso bloqueia publicação de tag quando a qualidade falha.

A validação local confirmou parser YAML, integrity guard e `git diff --check`. Esta sandbox não executa GitHub Actions, Docker distribuído, builds físicos macOS/Windows nem signing. SBOM/provenance/attestation continuam dependentes do runner e das credenciais/configurações de release; não foram declarados como concluídos localmente.


## Slice P1 de Grok/provider contract — 2026-09-22

O cliente Grok agora rejeita modelos fora de uma allowlist server-configured (`OLLAMA_AGENT_GROK_MODELS`, com o modelo padrão como fallback), e `Probe` só marca o provider como saudável quando o catálogo remoto contém pelo menos um modelo configurado. A rota HTTP `/api/agent/v1/grok/responses` rejeita `stream:true` com `501 Not Implemented` antes de chamar o upstream, evitando declarar streaming de produção quando o handler não expõe SSE; a API interna `StreamResponses` permanece testada para uma futura rota dedicada.

Testes cobrem modelo arbitrário, stream rejeitado, catálogo compatível/incompatível, allowlist por ambiente e não chamada ao upstream. Gates completos Go/integrity/build/UI/mobile passaram. Nenhuma credencial xAI foi usada ou declarada conectada; health real contra xAI continua dependente de credencial válida e ambiente autorizado.


## Slice P0 de Compose/infrastructure defaults — 2026-09-22

A composição agentic passou a exigir `OLLAMA_AGENT_POSTGRES_PASSWORD` e `OLLAMA_AGENT_REDIS_PASSWORD`, usar role PostgreSQL não-superuser, Redis com `requirepass` e portas PostgreSQL/Redis/OTLP ligadas a `127.0.0.1` por padrão. O init SQL não contém mais credencial fixa. O workflow distribuído cria passwords efêmeras por execução e injeta URLs autenticadas nos testes.

O integrity guard e `git diff --check` passaram. Docker/Compose não está instalado nesta sandbox; `docker compose config` e o teste real PostgreSQL/Redis/OTLP ficaram `NOT_RUN_DOCKER_UNAVAILABLE`, não sendo tratados como prova local. O GitHub Actions continua sendo a prova autoritativa dessa integração.


## Slice P1 de provider selection contract — 2026-09-22

A seleção de motor na Nova tarefa agora envia `provider` explicitamente e o runtime persiste essa escolha. Como somente o planner Ollama local está implementado nesta árvore, providers Claude/Codex/OmniRoute/automático aparecem como não conectados na UI e são rejeitados pelo servidor com erro de configuração, em vez de serem enviados silenciosamente ao Ollama como se fossem adapters reais.

O teste negativo server-side e os gates completos Go/integrity/UI passaram. A implementação de adapters externos, renovação OAuth e execução real dos CLIs continuam pendentes e não são declaradas conectadas.


## Slice P1 de release artifact integrity — 2026-09-22

O job final de release passou a rejeitar artefatos ausentes ou vazios antes de gerar checksums. A attestation de provenance pode ser executada somente quando a variável de ambiente GitHub `OLLAMA_ENABLE_ATTESTATIONS=true` estiver configurada, usando a permissão OIDC/attestations; caso contrário, permanece explicitamente não executada. A sandbox não publicou release, assinou binários ou executou Actions, portanto essas provas seguem externas.


## Slice P0 de tenant isolation em jobs/replay — 2026-09-22

Listagem e replay da fila agora são filtrados pela organização proprietária da missão. Um tenant não recebe jobs de outra organização e o replay cross-tenant retorna `403` sem alterar o estado da fila. A regressão HTTP foi adicionada ao teste combinado de escopo.

Gates completos Go/integrity/build/UI/mobile passaram. O isolamento ainda depende de prova distribuída com Redis/PostgreSQL reais e de revisão de outras superfícies de execução, mas jobs/replay não permanecem mais como leitura/mutação global no handler.


## Slice P0 de artifact manifest path safety — 2026-09-22

`BuildArtifactManifest` agora rejeita componentes symlink e resoluções fora do workspace antes de calcular hash, tamanho ou MIME. Foram adicionados testes para symlink externo e arquivo regular.

Gates completos Go/integrity/build/UI/mobile passaram. Isso protege a geração do manifest; a prova de publicação/exportação em todos os runtimes e adapters distribuídos continua separada.


## Remediação do workflow distribuído — 2026-09-22

O workflow `dz23-agentic-quality` executado no commit `2933f6e2` encontrou uma falha real no teste `TestDistributedPostgresRLSAndEvents`: a URL do smoke usava a role `ollama_agent` criada pelo Compose, que é a role bootstrap/superusuária, e o adapter tenant-scoped recusou corretamente operar com ela. O mesmo passo havia colocado passwords efêmeras no `GITHUB_ENV`, fazendo com que os valores aparecessem no bloco de ambiente do log do passo; esses valores eram efêmeros do CI, não credenciais de terceiros, e não foram adicionados ao repositório.

A correção publicada nesta iteração mantém as passwords somente em variáveis locais de um único passo, cria/ajusta uma role separada `ollama_agent_test` sem privilégio de superusuário e executa o smoke RLS usando essa role. O integrity guard também impede o retorno de `GITHUB_ENV` no job distribuído. A prova local desta correção é estática (parser YAML e integrity) e os gates completos locais passaram; a confirmação autoritativa depende de uma nova execução do GitHub Actions, porque Docker não está disponível nesta sandbox.

A classificação permanece **preview/local RC em hardening**. O CI distribuído continua sendo uma dependência de validação remota, não uma prova inventada localmente.


## Slice de workflow distribuído corrigida — 2026-09-22

- Falha observada no run `35735628693`: somente o teste PostgreSQL RLS falhou por uso de superusuário; Redis DLQ, OTLP e qualidade web/mobile passaram naquele run.
- Correção: role não-superusuária dedicada para o teste tenant-scoped e passwords confinadas ao passo de execução.
- Provas locais: integrity guard, YAML parser, `CGO_ENABLED=1 go test ./... -count=1`, `CGO_ENABLED=1 go vet ./...`, build Go, Vitest/build UI e typecheck mobile: PASS.
- Prova ainda pendente: novo run do GitHub Actions no commit desta correção.
- Não foram repetidas nem publicadas credenciais de terceiros; qualquer chave fornecida anteriormente continua considerada exposta e deve ser rotacionada pelo operador.


## Slice de workflow distribuído corrigida — 2026-09-22

O job PostgreSQL RLS deixou de conectar o teste tenant-scoped com a role bootstrap superusuária do Compose. O passo agora cria uma role de teste separada, executa toda a validação com ela e mantém passwords efêmeras em escopo local, sem `GITHUB_ENV`. O guard estático bloqueia regressão dessa prática.

Evidências locais desta correção: `scripts/check-class-a-plus-integrity.sh`, parser dos workflows, `git diff --check`, suíte Go completa com CGO, `go vet`, build Go, 20 arquivos/199 testes Vitest, build Vite e `npm run typecheck` mobile — todos PASS. Docker e a nova execução remota permanecem necessários para fechar o gate distribuído.

A classificação não muda: **preview/local RC em hardening**, sem merge automático em `main` e sem declarar produção-ready.
