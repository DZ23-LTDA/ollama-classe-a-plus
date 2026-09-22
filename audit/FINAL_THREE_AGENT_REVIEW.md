# Auditoria final de hardening — Ollama Classe A+

**Data:** 2026-09-22
**Escopo:** runtime agentic, shell desktop, Company OS/Growth OS, MCP/connectors, autenticação, mobile e gates de release.

## Conclusão

A rodada corrigiu uma camada adicional de riscos reproduzíveis em connectors, MCP, lifecycle administrativo e Browser Operator, mas a auditoria independente de 2026-09-22 encontrou **P0s internos ainda abertos** em isolamento multi-tenant, approvals, sandbox/process isolation, SSRF/DNS rebinding, DLP, Builder e release. O repositório possui uma base local-first funcional, com contratos, persistência, UI, smoke tests e guardrails parciais. O resultado correto é **preview/local RC em hardening**, não produto final nem produção-ready.

## Achados resolvidos

| Área | Correção verificada | Evidência |
|---|---|---|
| Autenticação | Bind não-loopback exige auth por padrão; `auth=false` não desliga essa proteção | `server/agent_auth_policy_test.go` |
| MCP stdio | Allowlist vazia é rejeitada e resposta com ID divergente falha | `internal/agent/mcp.go`, `internal/agent/mcp_test.go` |
| Remote MCP | HTTPS fora de loopback, allowlist, timeout, bloqueio de IP privado, redirect sem troca de origem e correlation ID | `internal/agent/mcp_remote.go`, `internal/agent/mcp_remote_test.go` |
| Configuração | Manifests MCP usam `DisallowUnknownFields` e rejeitam JSON residual | `server/agent_config_loader_test.go` |
| Filesystem | Componentes symlink e escapes de workspace são rejeitados; StepID é validado antes do sandbox | `internal/agent/tools.go`, `internal/agent/runtime_test.go` |
| Approvals | Approval registra tenant, policy, nonce, actor e expiração; decisão exige motivo e organização compatível | `internal/agent/runtime.go`, `internal/agent/runtime_test.go` |
| Growth OS | Mutações validam empresa/tenant; SKU é único; pedido suporta `Idempotency-Key`; fulfillment repetido é idempotente | `server/company_growth_routes.go`, `internal/agent/company_growth.go`, `company_growth_test.go` |
| Skills | Manifesto não pode marcar a própria skill como confiável sem atestado assinado futuro | `internal/agent/context.go` |
| UI | Settings e Control Center normalizam respostas nulas e preservam estados offline | 20 arquivos Vitest, 199 testes |
| Mobile | Outbox não persiste bearer; usa idempotency key, backoff, limite de tentativas e conflitos permanentes | `apps/mobile-agentic/App.tsx`, `npm run typecheck` |
| CLI/runtime | Teste antigo que rejeitava `ollama agent` foi corrigido; flags legados continuam rejeitados | `cmd/cmd_test.go` |
| Vet | Runtime não copia `sync.Mutex`; contexts de scheduler têm cancelamento em todos os retornos | `internal/agent/runtime.go`, `server/routes.go`, `go vet ./...` |

## Gates executados

Os seguintes comandos passaram no ambiente local:

```text
scripts/check-class-a-plus-integrity.sh
CGO_ENABLED=1 go test ./... -count=1
CGO_ENABLED=1 go vet ./...
CGO_ENABLED=0 go test ./internal/agent -count=1
CGO_ENABLED=1 go test ./server ./cmd/launch ./internal/multillm -count=1
CGO_ENABLED=1 go build -o /tmp/ollama-classe-a-plus-hardening .
app/ui/app: npm test -- --run       # 20 arquivos, 199 testes
app/ui/app: npm run build
apps/mobile-agentic: npm run typecheck
```

`CGO_ENABLED=0 go test ./...` não é um gate global aplicável a este fork porque os pacotes upstream de SQLite/MLX exigem CGO. O gate sem CGO continua válido para `internal/agent`, que é o núcleo independente de runtime agentic.

## Riscos residuais internos

A proteção de filesystem é uma contenção robusta contra symlinks presentes no momento da validação e deve evoluir para `openat`/handles de diretório por plataforma se o produto for exposto a concorrência hostil no mesmo filesystem. A política de skills ainda precisa de um formato de atestado assinado para promover uma skill a `trusted`; até lá, manifests carregados permanecem em revisão. Os testes de distributed RLS, Redis, OTLP, SAML/OIDC contra IdP real e companheiros físicos não podem ser simulados por testes locais.

## Blockers externos

Ainda faltam credenciais e ambientes autorizados para Composio, xAI, Desktop Commander Remote, Meta/Instagram, YouTube, WhatsApp, TikTok Shop, Shopify, CRM, anúncios e logística. Também faltam sandbox regional e app review de marketplaces; PostgreSQL/RLS, Redis e OTLP distribuídos; runners Windows/macOS/Linux; GPU e modelos multimídia locais; assinatura de instaladores; distribuição Android/iOS; e deploy real em contas do operador. Nenhum desses itens foi declarado conectado ou validado neste commit.

## Decisão de release

**Classificação:** `FIXING` / preview-local em hardening. A branch ainda não satisfaz os gates de conclusão porque existem P0 internos reproduzíveis, além dos blockers externos listados abaixo.
**Não classificar como produção universal:** os blockers internos e externos permanecem.
**Próxima ação segura:** publicar somente a camada verificada de hardening na PR pública, depois corrigir os P0 em slices verticais com testes negativos de tenant, approval, egress, DLP e Builder antes de qualquer classificação de conclusão.


## Achados da auditoria original que permanecem abertos

A correção desta rodada não encerra todos os 15 achados críticos originais. Permanecem abertos, por motivos técnicos ou externos, a prova de isolamento forte do sandbox com seccomp/cgroups/quotas; o ciclo completo Streamable HTTP/OAuth do MCP remoto, com sessão, refresh, resumption e pairing; uma capability policy assinada por skill com atestado de origem; auditoria e enforcement completos de egress para todos os connectors; storage seguro/bridge para tokens do desktop renderer; RLS e locking distribuído de produção; SBOM, provenance e assinatura verificável de artefatos; e smoke E2E com providers e dispositivos reais.

O **Growth OS** continua sendo um sandbox local reversível. Ele não é um agente comercial autorizado a publicar, comprar, cobrar, contratar, anunciar ou enviar pedidos. O **Desktop Commander** continua sendo um adapter com handshake e testes locais; nenhuma conta, máquina pareada ou credencial remota foi inventada. A matriz pública e o roadmap continuam sendo a fonte de verdade para esses gates posteriores.


## Incremento posterior à revisão

Depois da revisão inicial, foi adicionada uma policy de capabilities por missão. O default concede apenas escopos de workspace, enquanto browser, desktop, terminal, sandbox, MCP, connectors e deploy exigem escopos explícitos. Também foi removido o bypass genérico por sufixo `/connect`; somente o FullPath exato do handshake de dispositivo recebe tratamento especial, e o próprio handshake ainda exige transporte seguro, token de dispositivo e validações de sessão.

Os gates completos foram repetidos após esse incremento: suíte Go com CGO, `go vet`, build, Vitest, build UI, smokes funcionais e captura Chromium permaneceram verdes.


## Incremento operacional posterior — 2026-09-22

A rodada seguinte transformou adapters em superfícies operacionais locais. Grok Live ganhou cliente Responses com streaming, retry e circuito; Evaluation OS e provider router ganharam testes determinísticos; Company OS ganhou agentes departamentais e Social OS sandbox; plugins, MCP e skills ganharam lifecycle server-side; e o deploy Builder passou a checar approval antes de provider ausente.

As evidências adicionais foram `scripts/smoke-company-growth.sh`, `scripts/smoke-builder.sh`, `app/ui/app/scripts/smoke-shell.mjs`, captura Chromium das dez rotas do shell, `CGO_ENABLED=1 go test ./... -count=1`, `CGO_ENABLED=1 go vet ./...`, build Go, build UI, 20 arquivos/199 testes Vitest, typecheck mobile e integrity guard. O gate distribuído PostgreSQL/Redis/OTLP ficou `N/A` neste sandbox porque Docker não está instalado; não foi convertido em sucesso.

A classificação permanece `CANDIDATE_COMPLETED` para o release preview local. Os blockers externos continuam: contas e quotas xAI/Composio/Desktop Commander, OAuth e app review de redes sociais/marketplaces, staging distribuído, IdP, GPU, dispositivos físicos, assinatura, lojas e deploy do operador.


## Correções posteriores — 2026-09-22

A iteração final corrigiu riscos internos adicionais: `ConnectorManager.Call` não permite mais chamada sem organização; o caminho de operação usa matching por segmento; o dialer rejeita redes privadas resolvidas por DNS, liberando somente loopback explícito; Remote MCP e MCP stdio validam nomes de ambiente e Remote MCP rejeita headers de transporte; lifecycle global de plugins exige owner/admin quando autenticado; e o Browser Operator possui fallback de Chromium para caminhos configurados ausentes.

Provas: testes focados de connectors/MCP/server passaram; integrity guard passou; `CGO_ENABLED=1 go test ./... -count=1`, `CGO_ENABLED=1 go vet ./...`, build Go, Vitest 20/199, build UI, typecheck mobile e Browser Operator com caminho inválido passaram localmente. O CI remoto desta nova revisão ainda depende do workflow disparado após o commit.

## Estado após a auditoria independente ampliada — 2026-09-22

A auditoria ampliada foi considerada como evidência de risco, não como conclusão. Permanecem P0 internos que impedem declarar o produto final: isolamento por organização ainda não é demonstrado de forma uniforme para orchestration, traces, devices/pairing, Builder, plugins/MCP/skills e todas as operações CRUD; approvals ainda precisam de separação forte entre `approve` e `execute`, CAS/nonce e política de aprovador; o sandbox stdio/local continua sendo contenção best-effort sem seccomp/cgroups/rlimits/PID limits comprovados; Remote MCP/connectors/media ainda precisam de verificação do IP efetivamente conectado, redirect-chain e DLP uniforme; Company/Growth ainda aceitam invariantes server-managed e gasto/estado privilegiado que exigem correção atômica; Builder/artifacts precisam de ownership, entry validation e proteção contra symlink/XSS; release/CI ainda não prova o pacote de supply chain completo.

O próximo incremento deve começar por testes negativos cross-tenant e por DTOs allowlisted/ownership no servidor. Enquanto esses testes não passarem, `CANDIDATE_COMPLETED` não é uma classificação válida para o conjunto do produto. O Growth OS permanece sandbox local reversível e adapters externos permanecem não conectados sem credenciais, conta, dispositivo, sandbox ou evidência autorizada.


## Slice P0 validado — Builder e Company — 2026-09-22

O primeiro slice pós-auditoria ampliada adicionou `OrganizationID` aos projetos Builder e aplicou ownership server-side a todas as operações de projeto expostas pelos handlers, incluindo deploy e preview de arquivo. Também adicionou entry validation, recusa de symlink no preview/export e escaping de nomes em templates. No Company OS, o create usa DTO allowlisted e reseta campos server-managed; gasto de agente pausado ou acima do budget falha sem mutação do valor gasto.

A prova inclui `server/builder_scope_test.go`, que cria um projeto em `org-a` e demonstra listagem vazia e `403` sem mutação em `org-b`, além dos testes de entry/XSS/symlink e Company create/budget. Os gates completos do slice passaram. Isso encerra somente esta fronteira; orchestration, traces, devices/pairing, approvals fortes, sandbox/MCP isolation, egress/DLP e release supply chain continuam abertos.


## Slice P0 validada — orchestration, traces e devices — 2026-09-22

A segunda slice pós-auditoria adicionou `organization_id` aos orchestration jobs e aos spans de missão/ferramenta; os handlers de plan/get/run/cancel e traces globais passaram a filtrar pelo tenant. Devices e pairing aplicam ownership em listagem, heartbeat e revoke, e pairing codes não aceitam override de organização. O teste HTTP `server/p0_scope_test.go` cobre duas organizações e verifica `403` sem mutação para orchestration, traces e devices; testes de domínio cobrem pairing cross-tenant.

Os gates completos desta slice passaram. Isso não encerra os P0 da auditoria ampliada: plugins/MCP/skills/artifacts, approvals fortes, sandbox/process isolation, egress/DLP, Company Growth pausado, renderer session e release supply chain continuam abertos.


## Slice P0 validada — mission approvals — 2026-09-22

Approvals de missão agora separam a ação de aprovar do executor: em auth mode, owner/admin é obrigatório; a decisão valida organização, razão, nonce de uso único e versão corrente da missão. O Agentic Console envia o nonce retornado pela API. Testes cobrem role policy, nonce errado, replay e CAS; os gates completos passaram.

A correção não cobre ainda os booleans `approved` de Company/Growth/Social, que continuam uma lacuna explícita para decisão auditável, actor/policy/nonce/expiração e proteção contra auto-approval. Sandbox forte, egress/DLP e outros P0 também permanecem abertos.


## Slice P0 validada — CompanyApproval ledger — 2026-09-22

Campanhas, programas de afiliados, pedidos e drafts sociais passaram a usar approvals server-side com `resource_type/resource_id`, organization, policy, nonce, expiração, actor, razão e CAS de Company version. Os endpoints de aprovação não confiam mais no booleano ou no estado enviado pelo cliente; o campo `approved` é apenas uma projeção após decisão válida. A UI passa a enviar o nonce da decisão pendente.

A rota de gasto ainda possui `approved` booleano e não foi incluída nesta slice; orçamento, anúncios, contratos e mensagens externas exigem a continuação do approval ledger. A auditoria geral continua aberta.


## Slice P0 validada — spend approval e budget atômico — 2026-09-22

A rota HTTP de gasto não aceita mais `approved` como autoridade. Solicitações sujeitas à política criam uma decisão `spend` pendente e retornam `202` sem mutar `SpentCents`; a fila Company OS usa o endpoint genérico de decisão com owner/admin, nonce e CAS. A aprovação aplicada verifica pausa e limite antes do débito, e o teste HTTP/domínio cobre ausência de mutação antes da decisão e débito único depois dela.

O método interno legado com booleano continua somente para compatibilidade de domínio/testes e não é utilizado pela rota pública. A auditoria maior continua aberta para sandbox/MCP, egress/DLP, plugins e demais superfícies.


## Slice P0 validada — DLP de resultados e observabilidade — 2026-09-22

A redação recursiva agora é aplicada antes de persistir/emitir `Step.Result`, erros, eventos e atributos de traces, incluindo JSONStore e PostgresStore. Testes de token injection confirmam que credenciais em mapas, listas, eventos, traces e missões reabertas não aparecem em claro. A política de payload egress de connectors/MCP permanece aberta para não confundir redaction de dados com remoção indevida de credenciais operacionais autorizadas.


## Slice P0 validada — Remote MCP egress — 2026-09-22

O transporte Remote MCP desabilita proxy ambiental, limita redirects ao mesmo origin e confere o IP do socket efetivamente conectado após o dial. A regressão local confirma rejeição de endereço privado conectado e redirect same-origin permitido. Isso reduz DNS rebinding nessa superfície, mas não fecha Media/Connectors, TLS distribuído ou egress global.


## Slice P0 validada — MCP stdio lifecycle — 2026-09-22

MCP stdio agora aplica allowlist de executável absoluto não-symlink, cwd privado, ambiente mínimo, limites de payload e stderr, redaction e encerramento seguro de grupo com restart após cancelamento. Os testes negativos cobrem symlink, comando relativo, cleanup, timeout e payload excedente. A auditoria não considera isso sandbox forte: seccomp, cgroups, rlimits, PID/memória/CPU e validação em plataformas reais permanecem abertos.


## Slice P0 validada — Media egress/download — 2026-09-22

Downloads remotos de mídia passaram a bloquear redirects, proxy ambiental e destinos privados efetivamente conectados; respostas são bounded e só viram artefato após MIME/magic validation. Regressões cobrem redirect, MIME incompatível, magic inválido, overflow e IP privado. A paridade de egress para upload/Connectors, assim como testes distribuídos de DNS/TLS, ainda permanece aberta.


## Slice P0 validada — Connector egress parity — 2026-09-22

Connectors agora removem proxy ambiental, bloqueiam redirects, verificam o IP efetivamente conectado e limitam request/response. Testes negativos reproduzem redirect e payloads oversized em TLS local. A classificação de dados do usuário versus credenciais operacionais, uploads e validação distribuída permanecem abertas.


## Slice P0 validada — OAuth redirect URI allowlist — 2026-09-22

OAuth start/callback não aceitam mais redirect arbitrário: cada provider precisa de URI allowlisted e a forma canônica é usada no PKCE state. HTTPS é obrigatório, com loopback HTTP somente em opt-in explícito. A auditoria de egress dos endpoints OAuth e validação real contra um IdP/staging continuam abertas.


## Slice P0 validada — session handling web/mobile — 2026-09-22

O bearer web não é mais lido de localStorage: a sessão fica em memória e é limpa em 401/403. O mobile limpa SecureStore em sessão expirada, aplica nonce/busy/accessibility em approvals e pede confirmação antes de remover cache/outbox no logout. A jornada de login web e a validação física de secure storage nativo ainda não estão fechadas.


## Slice P0 validada — OAuth endpoint egress — 2026-09-22

OIDC discovery/JWKS/userinfo/token exchange agora usam egress sem proxy, redirects bloqueados e verificação de IP conectado; endpoints retornados por discovery são validados antes do uso. A prova contra IdP real, refresh/revogação e operação SSO de produção permanecem externas ao sandbox.


## Slice P0 validada — tool process containment — 2026-09-22

Terminal/sandbox tools ganharam grupo de processo encerrável, timeout, output bound, stderr redaction, ambiente mínimo e `ulimit` best-effort no sandbox. O resultado é explicitamente `best-effort`; seccomp/cgroups, quotas fortes e testes físicos por plataforma continuam abertos, portanto não equivale a sandbox forte.


## Slice P0 validada — plugin/MCP/skill ownership — 2026-09-22

Lifecycle autenticado de Connector, MCP stdio, Remote MCP e Skill agora exige ownership exato; catálogos preservam recursos sem owner como globais read-only. Testes negativos cobrem duas organizações e ausência de mutação cross-tenant. Registro tenant-owned server-side, attestation e isolamento forte de processos ainda permanecem abertos.


## Slice P0 validada — CI/release quality gates — 2026-09-22

Quality CI corrigiu indentação do Browser Operator e adicionou Go full/vet/build, UI Vitest/build e mobile typecheck. Release ganhou job de quality obrigatório para builds e publicação. Ações GitHub, Docker distribuído, runners físicos, signing e attestation ainda precisam de execução real; a alteração não os finge concluídos.
