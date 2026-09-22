# Integrações externas e MCP

## Conectores HTTP

O runtime aceita um arquivo JSON apontado por `OLLAMA_AGENT_CONNECTORS`. O arquivo contém apenas metadados, URLs HTTPS, operações permitidas e o nome da variável de ambiente que contém o token. Para tenants autenticados, uma operação pode declarar `oauth_provider` e o runtime resolve o access token cifrado da organização; nesse modo o token não é aceito pelo input da tool. O valor do token nunca entra no repositório, missão, memória, evento ou resposta de capabilities.

```bash
export OLLAMA_AGENT_CONNECTORS=/etc/ollama-dz23/agent-connectors.json
export DZ23_GITHUB_TOKEN='...'
export DZ23_GOOGLE_TOKEN='...'
export DZ23_SLACK_TOKEN='...'
export DZ23_DISCORD_TOKEN='...'
export DZ23_WHATSAPP_TOKEN='...'
```

O exemplo [agent-connectors.json](../../examples/agent-connectors.json) cobre GitHub, Google Workspace, Slack, Discord e WhatsApp Cloud. Cada chamada exige uma operação declarada, método permitido e prefixo de caminho permitido. Redirects são desativados, o timeout é limitado e a resposta é limitada a 2 MiB. Operações de escrita são tools de efeito externo e exigem approval do runtime.

O fluxo de autenticação OAuth usa PKCE, state one-time, armazenamento AES-GCM e refresh server-side. Configure os endpoints do provider e `OLLAMA_AGENT_CREDENTIAL_KEY` fora do repositório. Tokens pessoais por `token_env` continuam disponíveis apenas para desenvolvimento ou conectores explicitamente não multiusuário.

## SAML enterprise

O SP SAML é configurado por `OLLAMA_AGENT_SAML_<PROVIDER>_IDP_METADATA_URL`, `_METADATA_URL`, `_ACS_URL`, `_ENTITY_ID`, `_SP_PRIVATE_KEY_FILE`, `_SP_CERTIFICATE_FILE` e `_DEFAULT_REDIRECT_URI`. O metadata do IdP é validado pelo pacote `crewjam/saml`; o runtime assina AuthnRequests, usa RelayState one-time e valida o ACS antes de provisionar usuário e organização. O fluxo de primeiro login só fica público com `OLLAMA_AGENT_AUTH_SSO_PUBLIC=true`. Certificados, chaves e URLs devem ser provisionados pelo operador, e o teste ponta a ponta depende de um IdP real.

## MCP stdio

O runtime aceita um arquivo JSON apontado por `OLLAMA_AGENT_MCP`:

```json
[
  {
    "id": "filesystem-readonly",
    "command": "/usr/local/bin/my-mcp-server",
    "args": ["--readonly"],
    "allowed_methods": ["initialize", "tools/list", "tools/call"],
    "environment_vars": ["MCP_HOME"],
    "timeout_seconds": 30
  }
]
```

O processo é iniciado apenas quando uma chamada é feita. O manager usa JSON-RPC por stdin/stdout, limita methods, repassa somente variáveis explicitamente declaradas, desabilita shell e encerra o processo em timeout/cancelamento. O `mcp.call` é uma tool de efeito externo e exige approval.

O isolamento de filesystem e rede do servidor MCP ainda deve ser reforçado com um executor sandbox dedicado em instalações multiusuário. Nunca registre um MCP com `environment_vars` que incluam credenciais sem uma policy de tenant e auditoria equivalente.

## Desktop Commander local e Remote MCP

O [Desktop Commander](https://desktopcommander.app/) pode ser usado localmente como servidor MCP stdio com o preset [`examples/dz23-desktop-commander-mcp.json`](../../examples/dz23-desktop-commander-mcp.json). Ele requer Node.js 18+ e inicia o pacote `@wonderwhy-er/desktop-commander` com os métodos allowlisted. Filesystem, terminal, processos e edição continuam sujeitos ao registry de tools, approval e permissões do usuário do processo.

O Remote Desktop Commander oficial documenta o endpoint `https://mcp.desktopcommander.app/mcp`, Streamable HTTP, OAuth 2.0 com PKCE e pareamento por device flow. O preset [`examples/dz23-desktop-commander-remote.json`](../../examples/dz23-desktop-commander-remote.json) é carregado por `OLLAMA_AGENT_REMOTE_MCP`; o adapter usa HTTPS, allowlist de métodos, timeout e um bearer opcional em `DESKTOP_COMMANDER_ACCESS_TOKEN`. O runtime não tenta obter credenciais, abrir sessão OAuth, parear máquinas nem revogar dispositivos. A conta, o agente `npx ... remote`, a aprovação do código e a revogação devem ser feitos pelo operador no dashboard oficial.

O catálogo `GET /api/agent/v1/mcp` retorna servidores stdio e remotos sem o valor do token. Chamadas remotas passam pela tool `mcp.remote.call`, que exige approval. O serviço remoto está documentado como beta e executa com as permissões do usuário da máquina; não conecte uma conta de produção sem revisar escopos, dispositivo, logs e política de desligamento.

## HarnessRouter e harnesses de coding

O [HarnessRouter Community Edition](https://github.com/HarnessRouter/harnessrouter) pode ser configurado como provider `openai-compatible` em [`examples/dz23-harnessrouter.json`](../../examples/dz23-harnessrouter.json). Cada `ModelConfig` pode declarar `harness_id`; o proxy Classe A+ preserva metadata existente e sobrescreve `metadata.harness_id` no servidor, permitindo selecionar `harnessrouter/codex` ou `harnessrouter/claude-code` sem aceitar esse controle do browser.

A integração é opt-in, usa `HARNESSROUTER_API_KEY` somente no processo do servidor e mantém uma chave de entrada do gateway Classe A+ separada. O endpoint local HTTP é permitido apenas com allowlist de loopback; hosts externos exigem HTTPS. O adapter não prova instalação, autenticação, licença ou disponibilidade de um CLI: streaming, sessões de follow-up, cancelamento, artifacts e recovery precisam ser testados contra uma instância real antes de classificar o provider como validado.


## Infraestrutura distribuída

O runtime permanece local-first por padrão. Para persistência compartilhada, defina `OLLAMA_AGENT_DATABASE_URL` com uma URL PostgreSQL; o servidor executa migrações idempotentes para `agent_missions` e `agent_events`, preservando o JSON store como fallback quando a variável não existe. Para workers compartilhados, defina `OLLAMA_AGENT_REDIS_URL` e opcionalmente `OLLAMA_AGENT_REDIS_PREFIX`; a fila Redis implementa enqueue, claim, retry com backoff, dead-letter e replay. Não configure uma URL de produção com credenciais embutidas em arquivos versionados.

Para traces distribuídos, defina `OLLAMA_AGENT_OTLP_ENDPOINT` com uma URL HTTPS de OTLP HTTP. O provider exige HTTPS por padrão; `OLLAMA_AGENT_OTLP_ALLOW_INSECURE=1` é reservado para desenvolvimento local. A stack de desenvolvimento em `deploy/docker-compose.agentic.yml` fornece PostgreSQL, Redis e OpenTelemetry Collector.

## Companion WebSocket e mTLS

O endpoint `GET /api/agent/v1/devices/:id/connect` aceita WebSocket somente sobre TLS, salvo `OLLAMA_AGENT_ALLOW_INSECURE_COMPANION=1` para loopback local. `OLLAMA_AGENT_TLS_CERT_FILE` e `OLLAMA_AGENT_TLS_KEY_FILE` ativam TLS 1.3 no listener e recarregam o certificado por handshake; `OLLAMA_AGENT_REQUIRE_MTLS=1` exige também `OLLAMA_AGENT_TLS_CLIENT_CA_FILE`. O primeiro frame precisa ser `hello` com `device_id`, `token` e capabilities. `OLLAMA_AGENT_COMPANION_ORIGINS` limita Origins explícitas; a lista não deve ser `*` em produção.

O protocolo inicial é `dz23-companion.v1` e suporta `hello`, `heartbeat`, `ping` e respostas de rejeição para tipos ainda não habilitados. A camada de transporte não executa comandos arbitrários: ações de tela, teclado, processos e arquivos continuam sujeitas ao registry de tools, scopes e approvals do runtime.

## Publicação de builders

Configure `OLLAMA_AGENT_DEPLOYMENTS` apontando para um JSON como [`examples/agent-deployments.json`](../../examples/agent-deployments.json). Os adapters `vercel` e `netlify` usam as APIs oficiais; `generic` envia um payload de arquivos base64 para `/deploy`, permitindo integrar AWS, Cloudflare, um pipeline interno ou outro hosting sem colocar SDKs e credenciais no binário. Os tokens são lidos de `token_env`, e uma publicação exige approval no endpoint. O código cria a requisição de publicação e valida o workspace, mas credenciais de conta, domínio, projeto, DNS, billing e permissões de hosting continuam responsabilidade do operador.


## Composio Connect e plugin Composio

O Classe A+ pode registrar o [Composio Connect](https://docs.composio.dev/docs/composio-connect) como Remote MCP através de [`examples/dz23-composio-connect.json`](../../examples/dz23-composio-connect.json). O preset usa `https://connect.composio.dev/mcp`, permite somente os métodos JSON-RPC necessários (`initialize`, `notifications/initialized`, `tools/list` e `tools/call`) e injeta `x-consumer-api-key` apenas no servidor por meio de `COMPOSIO_CONSUMER_API_KEY`. O valor nunca é retornado pela API de capabilities, browser ou logs.

```bash
export OLLAMA_AGENT_REMOTE_MCP="$PWD/examples/dz23-composio-connect.json"
export COMPOSIO_CONSUMER_API_KEY='valor-fora-do-repositorio'
```

O Composio Connect expõe meta-tools para descobrir tools, obter schemas, iniciar conexões OAuth e executar tools. Por isso, “ter o plugin” significa ter o adapter MCP e o fluxo de aprovação no Classe A+; ainda é necessário autorizar cada conta upstream no navegador do operador. A integração não cria uma conta Composio, não completa OAuth automaticamente e não declara que Instagram, TikTok Shop, Shopify ou qualquer outro toolkit está conectado. Para multiusuário, a próxima evolução deve usar uma sessão Composio por `organization_id`/usuário, persistir somente referências cifradas e aplicar scopes mínimos por departamento.

## xAI / Grok por API

A API oficial da [xAI](https://docs.x.ai/overview) é OpenAI-compatible na Responses API. O preset [`examples/dz23-xai.json`](../../examples/dz23-xai.json) encaminha `/v1/responses` e `/v1/chat/completions` para `https://api.x.ai/v1`, usando `XAI_API_KEY` somente no processo do servidor:

```bash
export OLLAMA_DZ23_CONFIG="$PWD/examples/dz23-xai.json"
export OLLAMA_DZ23_GATEWAY_KEY='chave-do-cliente-fora-do-repositorio'
export XAI_API_KEY='chave-xai-fora-do-repositorio'
```

No cliente compatível com Responses API, use o modelo `xai/grok-4.7` e envie `input`, `tools` e as opções suportadas pela versão da API. O proxy Classe A+ preserva o corpo Responses, reescreve apenas o identificador lógico para o modelo upstream e aplica autenticação server-side. Isso integra a **API xAI**, não o produto hospedado Grok Bot. Browser, computador cloud persistente, bots coordenados, skills e rotinas continuam sendo implementados pelo runtime próprio do Classe A+ ou pelos adapters aprovados, sem copiar internals proprietários.

## Redes sociais, afiliados e marketplaces

A base atual tem Growth OS sandbox, conectores HTTP allowlisted, OAuth tenant-aware, MCP e approvals. Ela consegue planejar campanhas, criar rascunhos, manter catálogo/pedidos locais, registrar atribuição e simular fulfillment; não publica nem vende em uma plataforma externa sem um connector específico e credenciais autorizadas.

| Canal | Estado atual | Próximo adapter operacional |
|---|---|---|
| Instagram/Meta | Adapter genérico e catálogo Composio possível; publicação não validada no Classe A+ | Meta Login/OAuth, `instagram_business_content_publish`, mídia pública, webhooks, rate limit, approval e teste em conta profissional |
| X/Twitter | Toolkit Composio listado; não há conexão validada no projeto | OAuth, publicação/leitura permitida, rate limits, políticas de automação e approval |
| YouTube | Toolkit/API pública disponível; não há upload validado no projeto | OAuth Google, upload/resumable, metadata, quota, copyright e approval |
| WhatsApp | Connector HTTP/MCP possível; não há fluxo Cloud API validado | Meta Business, templates, opt-in, webhooks, proteção contra spam e approval |
| TikTok Shop | Growth sandbox; APIs oficiais cobrem catálogo, pedidos, fulfillment, promoções, finanças, webhooks e Affiliate Seller/Creator/Partner | App Partner Center, seller/creator authorization, scopes por região, sandbox, webhooks assinados, idempotência, returns/refunds, compliance e testes por mercado |
| Shopify | Pode ser conectado por Composio ou connector dedicado; não há OAuth/Admin GraphQL validado | App OAuth, scopes mínimos, produtos/pedidos, webhooks, rate limits e approval para mutações |
| Outros marketplaces | Connector genérico/MCP permite integração futura; nenhum marketplace é declarado conectado | Adapter específico por marketplace, catálogo, estoque, pedidos, logística, devoluções, pagamentos e reconciliação |

A documentação oficial consultada informa que as Affiliate APIs do TikTok Shop não estão disponíveis no Reino Unido e União Europeia e que o onboarding de creators não pode ser totalmente moderado por parceiros via API. Logo, a jornada “como TikTok Shop” é viável **arquiteturalmente** e pode ser implementada com autorização de seller/creator/partner, mas não pode ser marcada como operação real universal antes da aprovação do app, da região, dos escopos e da sandbox correspondente.

Nenhum connector deve publicar posts, iniciar anúncios, enviar mensagens, criar produtos, alterar preço/estoque, aprovar pedidos, solicitar fulfillment, cobrar ou movimentar dinheiro sem approval explícito, idempotency key, trilha de auditoria, limites de orçamento e política de pausa automática.
