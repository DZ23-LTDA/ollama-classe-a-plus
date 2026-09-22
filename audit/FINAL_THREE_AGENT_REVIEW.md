# Auditoria final de hardening — Ollama Classe A+

**Data:** 2026-09-22
**Escopo:** runtime agentic, shell desktop, Company OS/Growth OS, MCP/connectors, autenticação, mobile e gates de release.

## Conclusão

A rodada removeu os bloqueadores internos reproduzíveis identificados na auditoria consolidada anterior. O repositório possui uma base local-first funcional, com contratos, persistência, UI, smoke tests e guardrails de segurança. O resultado é um **candidato de release preview endurecido**, não uma declaração de que todas as contas externas, dispositivos, lojas e ambientes distribuídos estão validados.

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

**Classificação:** `CANDIDATE_COMPLETED` para hardening local e release preview.
**Não classificar como produção universal:** os blockers externos acima permanecem.
**Próxima ação segura:** publicar a branch verificada na PR pública, atualizar o checkpoint e manter a validação externa como jornadas separadas com credenciais, approval e rollback.


## Achados da auditoria original que permanecem abertos

A correção desta rodada não encerra todos os 15 achados críticos originais. Permanecem abertos, por motivos técnicos ou externos, a prova de isolamento forte do sandbox com seccomp/cgroups/quotas; o ciclo completo Streamable HTTP/OAuth do MCP remoto, com sessão, refresh, resumption e pairing; uma capability policy assinada por skill com atestado de origem; auditoria e enforcement completos de egress para todos os connectors; storage seguro/bridge para tokens do desktop renderer; RLS e locking distribuído de produção; SBOM, provenance e assinatura verificável de artefatos; e smoke E2E com providers e dispositivos reais.

O **Growth OS** continua sendo um sandbox local reversível. Ele não é um agente comercial autorizado a publicar, comprar, cobrar, contratar, anunciar ou enviar pedidos. O **Desktop Commander** continua sendo um adapter com handshake e testes locais; nenhuma conta, máquina pareada ou credencial remota foi inventada. A matriz pública e o roadmap continuam sendo a fonte de verdade para esses gates posteriores.


## Incremento posterior à revisão

Depois da revisão inicial, foi adicionada uma policy de capabilities por missão. O default concede apenas escopos de workspace, enquanto browser, desktop, terminal, sandbox, MCP, connectors e deploy exigem escopos explícitos. Também foi removido o bypass genérico por sufixo `/connect`; somente o FullPath exato do handshake de dispositivo recebe tratamento especial, e o próprio handshake ainda exige transporte seguro, token de dispositivo e validações de sessão.

Os gates completos foram repetidos após esse incremento: suíte Go com CGO, `go vet`, build, Vitest, build UI, smokes funcionais e captura Chromium permaneceram verdes.
