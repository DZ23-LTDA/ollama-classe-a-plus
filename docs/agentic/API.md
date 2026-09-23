# API agentic v1

A primeira API agentic roda no mesmo listener do Ollama e é local-first. Ela cria uma missão, gera um plano, aplica a policy de tools, persiste eventos e permite acompanhar o resultado.

## Configuração

Defina `OLLAMA_AGENT_ROOT` para o diretório que pode ser usado pelas missões. Por padrão, o runtime usa um diretório temporário do sistema. Defina `OLLAMA_AGENT_STORE` para o diretório persistente de missões, eventos e empresas. Se `OLLAMA_AGENT_MODEL` estiver definido, o runtime usa esse modelo para gerar o plano JSON; se o provider/modelo solicitado não tiver planner configurado ou a chamada falhar, a missão falha fechado em vez de usar um plano determinístico silencioso. `OLLAMA_AGENT_EMBED_MODEL` ativa memória semântica sobre a API `/api/embed`; `OLLAMA_AGENT_CONNECTORS`, `OLLAMA_AGENT_MCP` e `OLLAMA_AGENT_REMOTE_MCP` apontam para os manifestos de integração descritos em [INTEGRATIONS.md](INTEGRATIONS.md). `OLLAMA_AGENT_AUTH_STORE` habilita o store de identidade; `OLLAMA_AGENT_AUTH_REQUIRED=true` exige Bearer token; `OLLAMA_AGENT_CREDENTIAL_KEY` é obrigatório para persistir credenciais OAuth cifradas; cada provider OAuth pode declarar `OLLAMA_AGENT_OAUTH_<PROVIDER>_REVOCATION_URL`; `OLLAMA_AGENT_MEDIA_BASE_URL` e `OLLAMA_AGENT_MEDIA_API_KEY` ativam o adapter multimídia HTTPS.

```bash
export OLLAMA_AGENT_ROOT=/home/usuario/dz23-workspaces
export OLLAMA_AGENT_STORE=/home/usuario/dz23-data/agent-store
export OLLAMA_AGENT_MODEL=qwen3-coder:latest
export OLLAMA_AGENT_EMBED_MODEL=nomic-embed-text
export OLLAMA_AGENT_CONNECTORS=/etc/ollama-dz23/agent-connectors.json
export OLLAMA_AGENT_MCP=/etc/ollama-dz23/agent-mcp.json
export OLLAMA_AGENT_REMOTE_MCP=/etc/ollama-dz23/agent-remote-mcp.json
export OLLAMA_AGENT_SANDBOX_MODE=best-effort
# Em Linux multiusuário, aponte para um subtree cgroup v2 delegado ao processo:
# export OLLAMA_AGENT_SANDBOX_MODE=strict
# export OLLAMA_AGENT_SANDBOX_CGROUP_ROOT=/sys/fs/cgroup/ollama-agent
ollama serve
```

O modelo não recebe permissão implícita. Ele somente sugere passos; o runtime valida cada passo contra o registry e a policy.

### Sandbox de execução

`sandbox.exec` usa `best-effort` por padrão para preservar compatibilidade local. Em Linux, tenta user/mount/PID/network namespaces quando o host permite, além de `ulimit`, timeout, grupo de processo e saída limitada. Em macOS e Windows, executa pelo interpretador disponível como processo best-effort com timeout e saída limitada; o resultado informa `network_isolation=not-enforced`. Nenhum desses modos é uma fronteira forte de segurança.

Em Linux, `OLLAMA_AGENT_SANDBOX_MODE=strict` só executa quando `OLLAMA_AGENT_SANDBOX_CGROUP_ROOT` é um subtree cgroup v2 real e delegado com `cpu`, `memory` e `pids`. O runtime cria um cgroup efêmero por passo, aplica CPU, memória, PIDs e swap, coloca o processo nele por `CgroupFD`, usa namespaces sem rede, `setpriv --no-new-privs` e instala um filtro seccomp antes do interpreter. Se qualquer pré-requisito faltar — inclusive em macOS/Windows — a execução é recusada; nunca há downgrade silencioso para best-effort. O operador deve provisionar o subtree, validar a política do host e manter AppArmor/SELinux ou um executor/container dedicado quando a ameaça exigir uma fronteira mais forte.

## Criar missão

```bash
curl -sS http://localhost:11434/api/agent/v1/missions \
  -H 'Content-Type: application/json' \
	-d '{"objective":"inspecionar o workspace e listar os arquivos","capabilities":["workspace:read"],"auto_run":true}'
```

A resposta contém `mission_id`, estado, plano, approvals e timestamps. Uma missão de leitura pode entrar em execução automaticamente quando `auto_run` é verdadeiro.

`capabilities` é a lista de escopos concedidos à missão. Quando omitida, o runtime concede apenas `workspace:read` e `workspace:write` para preservar o modo local-first. Browser, desktop, terminal, sandbox, MCP, connectors e deploy exigem os escopos correspondentes de forma explícita. A aprovação continua necessária quando o descriptor da tool exigir.

## Consultar missão e eventos

```bash
curl -sS http://localhost:11434/api/agent/v1/missions/MISSION_ID
curl -sS http://localhost:11434/api/agent/v1/missions/MISSION_ID/events
curl -N http://localhost:11434/api/agent/v1/missions/MISSION_ID/events/stream
curl -sS http://localhost:11434/api/agent/v1/missions/MISSION_ID/traces
```

Eventos são append-only no store e servem como trilha de planejamento, início, retry, sucesso, bloqueio, aprovação e conclusão.

## Artefatos

Quando uma tool produz um arquivo, a missão registra nome, caminho relativo, MIME type, tamanho e SHA-256. O download exige o identificador da missão e do artifact; o runtime revalida que o arquivo ainda está dentro do workspace.

```bash
curl -OJ http://localhost:11434/api/agent/v1/missions/MISSION_ID/artifacts/ARTIFACT_ID
```

## Aprovar ação de escrita

Passos de escrita ou terminal são marcados com `requires_approval`. A missão não executa a tool antes da aprovação.

```bash
curl -sS -X POST \
  http://localhost:11434/api/agent/v1/missions/MISSION_ID/approvals/APPROVAL_ID \
  -H 'Content-Type: application/json' \
  -d '{"approved":true,"reason":"revisado pelo operador"}'

curl -sS -X POST http://localhost:11434/api/agent/v1/missions/MISSION_ID/run
```

A rejeição torna a missão falha de forma explícita. A aprovação não é reutilizada entre missões.

## Projetos e memória

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/projects \
  -H 'Content-Type: application/json' \
  -d '{"name":"Meu projeto","root":"/workspace"}'

curl -sS http://localhost:11434/api/agent/v1/projects
curl -sS -X PATCH http://localhost:11434/api/agent/v1/projects/PROJECT_ID \
  -H 'Content-Type: application/json' \
  -d '{"name":"Meu projeto atualizado","root":"/workspace"}'
curl -sS -X DELETE http://localhost:11434/api/agent/v1/projects/PROJECT_ID

curl -sS -X POST http://localhost:11434/api/agent/v1/projects/PROJECT_ID/memories \
  -H 'Content-Type: application/json' \
  -d '{"kind":"decision","content":"usar testes de contrato","confidence":1,"source":"operator"}'

curl -sS 'http://localhost:11434/api/agent/v1/projects/PROJECT_ID/memories?q=contrato'
```

Memórias são persistidas por projeto, têm fonte, confiança e, quando `OLLAMA_AGENT_EMBED_MODEL` está configurado, vetor embedding. A busca usa similaridade coseno e mantém fallback lexical quando o embedder não está disponível.

## Capabilities e observabilidade

```bash
curl -sS http://localhost:11434/api/agent/v1/tools
curl -sS http://localhost:11434/api/agent/v1/connectors
curl -sS http://localhost:11434/api/agent/v1/mcp
curl -sS http://localhost:11434/api/agent/v1/metrics
curl -sS http://localhost:11434/api/agent/v1/metrics/prometheus
```

O endpoint Prometheus expõe counters de missões, passos, retries, approvals e chamadas de tool.

`GET /api/agent/v1/connectors` retorna o catálogo tenant-aware e o estado seguro de cada connector. O campo `credential_configured` é apenas booleano; tokens, nomes de variáveis e ciphertext nunca fazem parte da resposta. Catálogo, adapter e credencial configurada não significam conta OAuth ativa: a conexão real ainda exige autorização do operador, teste reversível e evidência do upstream.

## Agendamentos e webhooks

Um schedule cria missões recorrentes. O intervalo é limitado a 31 dias e o claim é idempotente no store. `GET /schedules` lista os schedules do tenant ativo; `PATCH /schedules/SCHEDULE_ID` edita o objetivo, intervalo e estado; `DELETE /schedules/SCHEDULE_ID` remove a automação.

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/schedules \
  -H 'Content-Type: application/json' \
  -d '{"objective":"verificar o projeto","interval_seconds":3600,"webhook_secret_env":"DZ23_WEBHOOK_SECRET"}'

curl -sS http://localhost:11434/api/agent/v1/schedules
curl -sS -X PATCH http://localhost:11434/api/agent/v1/schedules/SCHEDULE_ID \
  -H 'Content-Type: application/json' \
  -d '{"objective":"verificar novamente","interval_seconds":7200,"enabled":true}'
curl -sS -X DELETE http://localhost:11434/api/agent/v1/schedules/SCHEDULE_ID
```

Para disparar por evento, configure a variável secreta no ambiente do processo e envie o header `X-Ollama-Agent-Secret`. O valor nunca é gravado no schedule.

```bash
export DZ23_WEBHOOK_SECRET='valor-fora-do-repositorio'
curl -sS -X POST http://localhost:11434/api/agent/v1/webhooks/SCHEDULE_ID \
  -H "X-Ollama-Agent-Secret: $DZ23_WEBHOOK_SECRET" \
  -H 'Content-Type: application/json' \
  -d '{"event":"push","ref":"main"}'
```

## Company OS

Uma empresa é criada no tenant ativo e persistida em `OLLAMA_AGENT_STORE/companies`. A criação inicial gera CEO/Estratégia, Produto, Engenharia, Marketing, Vendas, Suporte e Operações. Roadmap, metas, backlog e ciclos são dados da empresa, não fixtures da UI.

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/companies \
  -H 'Content-Type: application/json' \
  -d '{"name":"Minha empresa","mission":"Resolver um problema real","business_model":"SaaS","budget":{"currency":"BRL","monthly_limit_cents":100000,"approval_threshold_cents":10000}}'

curl -sS http://localhost:11434/api/agent/v1/companies
curl -sS http://localhost:11434/api/agent/v1/companies/COMPANY_ID/report
curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/backlog \
  -H 'Content-Type: application/json' \
  -d '{"title":"Validar oferta com clientes","priority":10,"owner_department":"sales"}'
curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/cycles \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ciclo diário","objective":"Revisar métricas e executar o próximo item","frequency":"daily","interval_seconds":86400}'
```

Criar um ciclo também cria um schedule persistente com workspace `company://COMPANY_ID`. `POST /companies/COMPANY_ID/pause` impede que o worker crie novas missões para os ciclos dessa empresa. Gasto em anúncios/contratos ou acima do limiar exige `approved:true`; limite excedido e anomalias de alta severidade pausam a empresa.

O Growth OS local usa approval explícito antes de qualquer efeito externo:

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/campaigns \
  -H 'Content-Type: application/json' \
  -d '{"name":"Campanha de validação","channel":"social","objective":"Gerar leads","daily_budget_cents":0}'
curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/campaigns/CAMPAIGN_ID/approve
curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/campaigns/CAMPAIGN_ID/launch

curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/products \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU-001","name":"Produto sandbox","supplier":"Fornecedor sandbox","cost_cents":500,"price_cents":1200,"inventory":10}'
curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/orders \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"PRODUCT_ID","customer_ref":"customer-test","quantity":1}'
curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/orders/ORDER_ID/approve
curl -sS -X POST http://localhost:11434/api/agent/v1/companies/COMPANY_ID/orders/ORDER_ID/fulfill \
  -H 'Content-Type: application/json' -d '{"tracking_code":"SANDBOX-001"}'
```

O endpoint `GET /companies/COMPANY_ID/growth/report` resume campanhas ativas, conversões de afiliados, catálogo, pedidos e receita registrada. A implementação não chama redes sociais, redes de afiliados, marketplaces, fornecedores, gateways de pagamento ou transportadoras.

## Tools da primeira fatia

`workspace.list` lê entradas do workspace autorizado e limita a quantidade retornada. `workspace.read` lê no máximo 1 MiB e rejeita traversal. `workspace.write` exige approval, limita o payload, escreve com permissões restritas e cria um manifesto com tamanho e SHA-256. `terminal.exec` existe em modo conservador e só permite `pwd`, `ls` e `git status` com argumentos separados; shell interpolation e outros executáveis são bloqueados. `sandbox.exec` executa Python ou Node. Em Linux best-effort tenta namespaces; em macOS/Windows usa processo best-effort e declara que a rede não está isolada. O modo strict exige Linux, cgroup v2 delegado e approval. `browser.operator` usa Chromium/Playwright com perfil persistente por sessão, bloqueio SSRF, snapshot, click, fill, press, upload, download, screenshot e estado `approval_required` para takeover humano. `desktop.companion` possui adapters Linux, macOS e Windows para screenshot, mouse, teclado, clipboard e processos através de executáveis separados e sem shell concatenado. `mcp.call` usa JSON-RPC stdio com lifecycle, timeout, env allowlist e methods allowlisted. `mcp.remote.call` usa Remote MCP Streamable HTTP, exige HTTPS fora de loopback, aceita bearer apenas por variável de ambiente e mantém allowlist de métodos; OAuth PKCE e pareamento do provedor são externos. `connector.http` chama GitHub, Google, Slack, Discord ou WhatsApp somente por operação HTTPS declarada.

### Sandbox strict

`OLLAMA_AGENT_SANDBOX_MODE=strict` habilita a fronteira Linux com cgroup v2 delegado, namespaces, `no-new-privs`, seccomp e limites de CPU/memória/PIDs. Se `OLLAMA_AGENT_SANDBOX_CGROUP_ROOT` não apontar para um cgroup v2 delegado com os controllers necessários, a execução falha fechado. O modo vazio ou `best-effort` não deve ser tratado como isolamento forte; AppArmor/SELinux e homologação do host continuam responsabilidade do operador.

## OAuth e organizações

Com `OLLAMA_AGENT_AUTH_REQUIRED=true`, a API valida sessão, organização e RBAC antes de permitir ações. O fluxo `GET /auth/oauth/:provider/start` exige `redirect_uri` e `code_verifier` PKCE; o callback consome `state` uma única vez, troca o código no servidor e cifra access/refresh tokens com AES-GCM usando `OLLAMA_AGENT_CREDENTIAL_KEY`. Configure URLs e nomes das variáveis de segredo por provider sem colocar valores no repositório.

Mutations autenticadas com `Origin` usam a allowlist de `OLLAMA_ORIGINS`; origens cross-site são recusadas antes do Bearer/RBAC. `GET`, `HEAD`, `OPTIONS` e clientes nativos sem `Origin` não são bloqueados por esse guard, mas continuam sujeitos a autenticação e autorização. O bypass público de SSO fica limitado a start/callback/metadata/ACS e não alcança refresh, revoke, logout ou ações de tenant.

Depois do callback, o cliente pode renovar e revogar uma credencial somente dentro da organização autenticada:

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/auth/oauth/github/refresh \
  -H 'Authorization: Bearer TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{"credential_id":"CREDENTIAL_ID"}'

curl -i -X POST http://localhost:11434/api/agent/v1/auth/oauth/github/revoke \
  -H 'Authorization: Bearer TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{"credential_id":"CREDENTIAL_ID"}'
```

O refresh faz rotação quando o provider retorna um novo refresh token, usa CAS para impedir duas renovações concorrentes e não retorna ciphertext. A revogação chama o endpoint do provider quando `REVOCATION_URL` está configurado e sempre persiste a revogação local após uma resposta 2xx. O bypass público de SSO não se aplica a refresh ou revoke.

Para encerrar uma sessão local, envie o bearer atual a `POST /api/agent/v1/auth/logout`. O servidor revoga o token e responde `204`; o cliente deve limpar sua cópia em memória mesmo se a chamada falhar. `403` indica escopo/RBAC recusado e não deve ser tratado como expiração automática da sessão.

```bash
curl -i -X POST http://localhost:11434/api/agent/v1/auth/logout \
  -H "Authorization: Bearer $OLLAMA_AGENT_ACCESS_TOKEN"
```

### SAML enterprise

`GET /auth/saml/:provider/start` cria AuthnRequest assinado e retorna `authorization_url` e RelayState one-time. `GET /auth/saml/:provider/metadata` publica o metadata do SP. `POST /auth/saml/:provider/acs` valida assinatura, audiência, destination, condições e `InResponseTo` pelo adapter `crewjam/saml`, extrai claims e cria sessão local tenant-aware. O fluxo exige `OLLAMA_AGENT_AUTH_SSO_PUBLIC=true` quando usado como primeiro login público.

As variáveis são `OLLAMA_AGENT_SAML_<PROVIDER>_IDP_METADATA_URL`, `_METADATA_URL`, `_ACS_URL`, `_SP_PRIVATE_KEY_FILE`, `_SP_CERTIFICATE_FILE`, `_ENTITY_ID` e `_DEFAULT_REDIRECT_URI`. URLs de metadata, SP e ACS devem ser HTTPS; certificados e chaves devem ser montados fora do repositório.

### TLS, mTLS e RLS

`OLLAMA_AGENT_TLS_CERT_FILE` e `OLLAMA_AGENT_TLS_KEY_FILE` ativam TLS 1.3 para o listener. `OLLAMA_AGENT_REQUIRE_MTLS=1` exige também `OLLAMA_AGENT_TLS_CLIENT_CA_FILE` e valida certificado de cliente. O certificado do servidor é recarregado a cada handshake, permitindo rotação sem reiniciar o processo. Quando o store PostgreSQL é usado, `FORCE ROW LEVEL SECURITY` e o contexto transacional `app.current_organization_id` impedem acesso entre organizações, inclusive para o dono da tabela.

## Mídia

Com um provider HTTPS compatível, use `POST /media/image`, `/media/video`, `/media/speech` e `/media/transcribe`, sempre com `mission_id`; os resultados são salvos em `.agent-media` e devolvem manifesto com hash. `POST /media/tone` gera um WAV determinístico local para smoke tests e não deve ser confundido com geração musical neural.

## Builders e preview

`POST /builders` cria `website`, `app`, `game`, `slides` ou `dashboard` a partir de arquivos controlados ou template. `POST /builders/:id/preview` valida a entrada e retorna manifesto; `GET /builders/:id/preview/*path` serve preview local com containment; `POST /builders/:id/export` cria ZIP; `POST /builders/:id/publish` publica uma cópia versionada localmente. Deploy público exige um adapter de hosting configurado e não é alegado por essa publicação local.

`POST /builders/:id/visual` atualiza componentes, bindings, estilos e eventos do canvas. `POST /builders/:id/undo` e `POST /builders/:id/redo` alteram o histórico persistido, incrementam a versão e regeneram o preview. O histórico é limitado às últimas 50 alterações para impedir crescimento sem limite.

`GET /deployments` lista somente os providers configurados sem tokens. `POST /builders/:id/deploy/:provider` empacota o workspace contido do projeto e chama Vercel, Netlify ou um deployer genérico declarado em `OLLAMA_AGENT_DEPLOYMENTS`. O corpo precisa conter `{"approved":true}`; sem essa aprovação explícita a API retorna `428 Precondition Required`. O token é lido exclusivamente de `token_env` no servidor. O adapter impõe limite de 2.000 arquivos, 50 MiB, redirects desabilitados e HTTPS fora de loopback.

## Colaboração

`GET /collab/:project_id`, `POST /collab/:project_id/comments`, `POST /collab/:project_id/presence` e `GET /collab/:project_id/stream` oferecem snapshot persistente, comentários, presença e SSE. Quando autenticação está ativa, o actor vem da sessão e o tenant é validado pelo middleware.

## Mobile

O cliente Expo em `apps/mobile-agentic` usa SecureStore para o Bearer token, polling compatível com Android/iOS e EAS profiles para development/preview/production. Assinatura, credenciais de loja, `ascAppId` e hosting da API continuam fora do código.


## Orquestração multiagente

`POST /api/agent/v1/orchestration/jobs` recebe `objective`, `roles` opcionais (`research`, `programming`, `testing`, `design`, `security`, `data`, `review`), `workspace`, `project_id` e `budget`. Se `auto_run` for verdadeiro, o job começa em background; caso contrário, a API devolve um job `PLANNED`.

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/orchestration/jobs \
  -H 'Content-Type: application/json' \
  -d '{"objective":"pesquisar e revisar a arquitetura","roles":["research","security","review"],"budget":{"max_agents":3,"max_seconds":600,"max_retries":1},"auto_run":true}'

curl -sS http://localhost:11434/api/agent/v1/orchestration/jobs/ORCH_ID
curl -sS -X POST http://localhost:11434/api/agent/v1/orchestration/jobs/ORCH_ID/run
curl -sS -X POST http://localhost:11434/api/agent/v1/orchestration/jobs/ORCH_ID/cancel
```

O orquestrador limita concorrência, tempo, tamanho de saída e retries por job. A síntese é determinística e conserva a saída de cada papel, citações e conflitos de evidência; um modelo posterior pode substituir o reducer sem remover a validação do runtime.

## Pesquisa profunda

`POST /api/agent/v1/research` baixa URLs HTTPS públicas em paralelo, limita bytes por fonte, aplica bloqueio SSRF, respeita `robots.txt` quando solicitado, usa cache por URL e retorna fontes, SHA-256, texto extraído, citações e erros individuais.

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/research \
  -H 'Content-Type: application/json' \
  -d '{"query":"arquitetura agentic","urls":["https://example.org"],"max_sources":8,"respect_robots":true}'
```

A pesquisa não acessa hosts privados, loopback ou link-local em produção. Login, CAPTCHA e fontes autenticadas continuam exigindo o fluxo Browser Operator com takeover humano; o ResearchEngine não tenta contornar autenticação.

## Dispositivos e companions

`POST /devices/pair/start` cria um código one-time com expiração curta. O companion envia esse código em `POST /devices/pair/complete` junto com plataforma, nome e capability report; o token retornado deve ser armazenado somente no dispositivo. `POST /devices/:id/heartbeat` aceita Bearer token ou `X-Device-Token` e atualiza capabilities. `GET /devices` lista somente metadados; `POST /devices/:id/revoke` revoga o token e impede novos heartbeats.

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/devices/pair/start
curl -sS -X POST http://localhost:11434/api/agent/v1/devices/pair/complete \
  -H 'Content-Type: application/json' \
  -d '{"pairing_code":"CODE","name":"Meu Desktop","platform":"darwin","capabilities":[{"name":"screen","scopes":["desktop:screen"]}]}'
```

O pairing local não substitui mTLS, assinatura de binário ou store distribuído de produção. Esses gates permanecem obrigatórios antes de expor companions em rede pública.


## Ingestão e memória documental

`POST /projects/:id/ingest` importa arquivos do root autorizado ou URLs públicas para chunks de memória. O ingestor suporta texto, Markdown, HTML, JSON, CSV, PDF via `pdftotext`, DOCX via `document.xml` e XLSX via worksheets; cada chunk conserva o caminho/URL de origem. Embeddings são produzidos pelo embedder configurado no ContextStore. Imagens exigem um adapter de visão/OCR configurado e não são tratadas como texto automaticamente.

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/projects/PROJECT_ID/ingest \
  -H 'Content-Type: application/json' \
  -d '{"paths":["docs/spec.pdf","README.md"],"chunk_size":1800,"chunk_overlap":200}'
```


## Grok Live, Evaluation OS e routing

`GET /api/agent/v1/grok/status` retorna o estado sanitizado do provider xAI/Grok. `POST /api/agent/v1/grok/responses` encaminha uma requisição Responses com `input`, `model`, `tools` e `stream` usando a chave configurada somente no servidor. Sem credencial, o endpoint não tenta uma chamada externa.

O Evaluation OS executa casos determinísticos de coding, browser, tools, segurança, memória, planejamento e recuperação. O provider router pode restringir uma decisão a modelos locais, exigir capacidades, impor orçamento e ordenar candidatos por saúde, latência, custo e qualidade histórica. Essas métricas não provam qualidade geral nem substituem avaliação com dados autorizados.

## Lifecycle de plugins

Os endpoints seguintes alteram somente recursos já registrados no servidor e são protegidos pelo mesmo middleware agentic:

- `POST /api/agent/v1/connectors/:id/enable` e `POST /api/agent/v1/connectors/:id/disable`;
- `POST /api/agent/v1/mcp/:id/enable` e `POST /api/agent/v1/mcp/:id/disable`;
- `POST /api/agent/v1/remote-mcp/:id/enable` e `POST /api/agent/v1/remote-mcp/:id/disable`;
- `POST /api/agent/v1/skills/:id/enable` e `POST /api/agent/v1/skills/:id/disable`;
- `DELETE` nos recursos correspondentes para remoção explícita.

A UI não concede scopes por conta própria. O servidor valida o identificador, o estado e as allowlists antes de alterar o lifecycle. Tokens continuam fora das respostas e manifests sem atestado permanecem sem confiança executável.


## Lifecycle de connectors

`GET /api/agent/v1/connectors` lista somente connectors acessíveis à organização ativa e redige `token_env`; `credential_configured` é calculado no servidor como booleano. Quando `OLLAMA_AGENT_CONNECTORS` não está definido, o runtime usa o manifest durável `OLLAMA_AGENT_STORE/connectors.json`. O arquivo é criado com diretório privado, modo `0600` e atualização por arquivo temporário seguido de rename. Ele contém referências de configuração, nunca valores de token, ciphertext ou headers secretos.

Em modo autenticado, `POST /api/agent/v1/connectors` exige membership `owner` ou `admin`. O servidor força `organization_id` a partir do contexto autenticado. Um `organization_id` conflitante, um ID já pertencente a outro tenant, campos JSON desconhecidos e valores de segredo como `token` ou `api_key` são rejeitados. O payload usa apenas nomes de variáveis de ambiente e/ou identificadores de provider OAuth:

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/connectors \
  -H 'Authorization: Bearer TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{"id":"github-prod","provider":"GitHub","base_url":"https://api.github.com","token_env":"DZ23_GITHUB_TOKEN","oauth_provider":"github","operations":[{"name":"profile","methods":["GET"],"path_prefixes":["/user"]}]}'
```

O endpoint valida HTTPS sem userinfo, nomes de ambiente, operações, métodos e prefixos de caminho. Requests e responses do connector permanecem limitados pelo manager e o egress mantém bloqueio de proxy/redirect e verificação do endereço conectado. A configuração `OLLAMA_AGENT_CONNECTORS=/caminho/manifest.json` continua disponível como modo estático explícito; nesse modo o manifest é carregado na inicialização e o lifecycle cadastrado pela UI não é apresentado como persistente nesse arquivo. O operador deve escolher conscientemente entre o manifest estático e o store durável.

A resposta do catálogo nunca devolve o nome da variável de token. Registrar um endpoint não prova que a conta do provider existe, que OAuth foi consentido, que o upstream está saudável ou que uma ação externa foi executada. Esses estados exigem credencial provisionada fora do repositório, escopos mínimos, approval e smoke reversível autorizado.


## Registro de MCP, Remote MCP e skills

Os endpoints de registro seguem o mesmo contrato de `POST /connectors`: autenticação owner/admin, binding server-side da organização, `DisallowUnknownFields` e resposta de catálogo redigida. O MCP stdio aceita `id`, `command` absoluto, `args`, `working_directory`, `allowed_methods`, `environment_vars` e `timeout_seconds`. O servidor exige arquivo executável regular, rejeita symlink como executable, limita args e valida nomes de env. Exemplo:

```json
{"id":"desktop-commander","command":"/opt/agent/bin/mcp-server","args":["--stdio"],"allowed_methods":["tools/list","tools/call"],"environment_vars":["DESKTOP_COMMANDER_TOKEN"]}
```

Remote MCP aceita `id`, `url`, `token_env`, `headers_env`, `allowed_methods` e `timeout_seconds`. A URL deve ser HTTPS fora de loopback, não pode conter userinfo/fragment/private address, e redirects precisam conservar origin permitido. Os valores de token/header nunca entram no JSON; somente nomes de env são armazenados. O DNS é resolvido e os endereços aprovados são pinados para a conexão.

Skills aceitam `id`, `version`, `description`, `scopes` e `tools`. `trusted` e `enabled` são estados derivados do servidor; o request com qualquer um desses estados como `true` é rejeitado. O runtime persiste o manifest em `OLLAMA_AGENT_STORE/context/skills/<id>.json`, sempre com `trusted=false`, para permitir revisão/approval explícita.

Sem `OLLAMA_AGENT_MCP` ou `OLLAMA_AGENT_REMOTE_MCP`, os managers padrão usam `OLLAMA_AGENT_STORE/mcp.json` e `remote-mcp.json`. Quando o operador define um arquivo estático por variável de ambiente, ele é carregado como bootstrap e mutations posteriores ficam apenas no manager da execução; a API não afirma ter editado o arquivo externo. Registrar é uma operação de configuração local: não equivale a OAuth consentido, conta conectada, smoke de upstream ou ação externa executada.


## Logout e escopo de sessão

`POST /api/agent/v1/auth/logout` tem dois comportamentos explícitos. Quando `auth_required=false`, é uma operação local idempotente e responde `204 No Content` mesmo sem bearer; não há sessão remota para revogar. Quando `auth_required=true`, o middleware exige bearer válido e o handler marca o hash do token como revogado, respondendo `204`. Tokens ausentes, inválidos ou expirados continuam retornando `401` no modo autenticado.

Esse endpoint revoga somente o token do runtime agentic. Não revoga automaticamente tokens OAuth de connectors, sessões de IdP ou contas de terceiros; cada credential/provider possui seu fluxo separado de refresh/revoke e requer homologação externa.


## Origin policy

Mutations agentic com `Origin` presente são aceitas somente quando a origem está em `OLLAMA_ORIGINS` ou nos defaults loopback (`localhost`, `127.0.0.1`, `0.0.0.0`, HTTP/HTTPS e portas). Essa política é aplicada com ou sem bearer. GET/HEAD, OPTIONS e clientes nativos sem header `Origin` preservam o caminho compatível. A regra é uma proteção de navegador e não substitui autenticação ou mTLS.


## Allowlist de Origin com portas variáveis

Para permitir portas variáveis em um host, use `https://host.example:*`. Esse formato compara scheme e hostname exatos e exige uma porta no header `Origin`; ele não permite subdomínios parecidos. Wildcards de esquema como `app://*` seguem sendo explícitos e separados. Userinfo, path, query e fragment não são aceitos no formato de porta variável.


## Bodies JSON

Os handlers agentic que usam `decodeJSON` aceitam exatamente um documento JSON. Campos desconhecidos são rejeitados e qualquer segundo objeto, array ou conteúdo inválido após o primeiro produz `400`. Whitespace final é permitido. Essa regra é complementar à autorização, approval e validação específica do endpoint.


O limite de transporte do `decodeJSON` é 4 MiB por request. Bodies maiores são rejeitados antes de concluir o parse; endpoints com payloads específicos podem impor limites menores. O orçamento não substitui limites de arquivos, documentos comprimidos ou respostas de providers.


`POST /api/agent/v1/auth/dev/token` é apenas uma rota de desenvolvimento: requer `OLLAMA_AGENT_AUTH_DEV=true` e peer loopback. Seu body também respeita o limite de 4 MiB, campos conhecidos e documento JSON único; a rota não deve ser exposta como onboarding de produção.


## Deployments

O adapter de deployment aceita HTTPS para serviços externos e HTTP somente em loopback. O root do workspace não pode ser symlink; redirects são bloqueados; a conexão padrão verifica o endereço IP real após o DNS e rejeita destinos privados não-loopback. O smoke local usa um servidor fixture. Configurar Vercel, Netlify, AWS, Cloudflare ou outro serviço não significa conta validada, domínio publicado, billing autorizado ou rollback testado.


## Mídia

`Transcribe` e `AnalyzeImage` exigem arquivos regulares dentro do workspace informado. O workspace final e os componentes do caminho não podem ser symlinks; traversal e input externo são rejeitados. Transcrição acima de 100 MiB falha antes da leitura. Geração e speech gravam somente em um workspace validado. Provider fixture local valida contrato, não conta, quota, GPU ou upstream externo.


## Company Growth OS

Campaigns, affiliate programs/links e orders retornam `mode: "sandbox"`; o relatório de crescimento retorna `sandbox_only: true`. O runtime pode planejar, aprovar, registrar métricas, manter inventário e simular fulfillment local. Um modo diferente de sandbox retorna `external growth action is not configured` até existir adapter, connector, credencial e smoke autorizado. Esses endpoints não cobram, publicam, movimentam dinheiro, fazem pedido externo ou emitem NF-e.


## Configuração segura

`GET /api/agent/v1/config` considera tanto arquivos/env de bootstrap quanto manifests registrados nos managers duráveis. `connectors_configured`, `mcp_configured`, `media_configured` e `deployments_configured` indicam configuração aceita pelo Runtime. Esses booleanos não afirmam credencial presente, conta OAuth conectada, resposta upstream, publicação ou homologação.


## Contratos de mídia e deploy após a reauditoria

Outputs de mídia e OCR são escritos somente por paths relativos ao workspace autorizado. A implementação usa `os.Root` e rejeita raiz, diretório ou componente symlink; uma troca concorrente para um alvo fora do root falha fechada. A validação do diretório de output ocorre antes de chamar provider ou processo externo. A leitura de áudio e imagem abre o descritor dentro do root e aplica orçamento cancelável. `Transcribe` rejeita entradas acima de 100 MiB. `AnalyzeImage` rejeita entradas acima de 25 MiB antes da leitura integral e também falha se o arquivo crescer durante a leitura.

O coletor de deployment aceita somente arquivos regulares de conteúdo público. `.git`, `.hg`, `.svn`, `.agent`, `.ollama`, `.secrets`, `node_modules`, `.env`, chaves, certificados, backups, dumps, logs e nomes que indicam secret/credential/password/token são excluídos antes da leitura. Isso é uma política conservadora de pacote; um build que necessite de conteúdo adicional deve produzir um diretório público aprovado, não depender de filtro posterior.

O cliente de deployment resolve A/AAAA antes do TCP. Todos os endereços retornados precisam ser não privados antes que o socket seja criado. O dial usa o IP validado, enquanto a URL mantém hostname para Host/SNI. HTTP é aceito somente para loopback em contexto local explícito; endpoints externos exigem HTTPS e redirects continuam desabilitados.
