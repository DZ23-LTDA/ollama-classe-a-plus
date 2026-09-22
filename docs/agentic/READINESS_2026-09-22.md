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
