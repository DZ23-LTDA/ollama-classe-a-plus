# API agentic v1

A primeira API agentic roda no mesmo listener do Ollama e é local-first. Ela cria uma missão, gera um plano, aplica a policy de tools, persiste eventos e permite acompanhar o resultado.

## Configuração

Defina `OLLAMA_AGENT_ROOT` para o diretório que pode ser usado pelas missões. Por padrão, o runtime usa um diretório temporário do sistema. Defina `OLLAMA_AGENT_STORE` para o diretório persistente de missões e eventos. Se `OLLAMA_AGENT_MODEL` estiver definido, o runtime tenta usar esse modelo para gerar o plano JSON; em caso de erro, usa o planner determinístico seguro.

```bash
export OLLAMA_AGENT_ROOT=/home/usuario/dz23-workspaces
export OLLAMA_AGENT_STORE=/home/usuario/dz23-data/agent-store
export OLLAMA_AGENT_MODEL=qwen3-coder:latest
ollama serve
```

O modelo não recebe permissão implícita. Ele somente sugere passos; o runtime valida cada passo contra o registry e a policy.

## Criar missão

```bash
curl -sS http://localhost:11434/api/agent/v1/missions \
  -H 'Content-Type: application/json' \
  -d '{"objective":"inspecionar o workspace e listar os arquivos","auto_run":true}'
```

A resposta contém `mission_id`, estado, plano, approvals e timestamps. Uma missão de leitura pode entrar em execução automaticamente quando `auto_run` é verdadeiro.

## Consultar missão e eventos

```bash
curl -sS http://localhost:11434/api/agent/v1/missions/MISSION_ID
curl -sS http://localhost:11434/api/agent/v1/missions/MISSION_ID/events
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

curl -sS -X POST http://localhost:11434/api/agent/v1/projects/PROJECT_ID/memories \
  -H 'Content-Type: application/json' \
  -d '{"kind":"decision","content":"usar testes de contrato","confidence":1,"source":"operator"}'

curl -sS 'http://localhost:11434/api/agent/v1/projects/PROJECT_ID/memories?q=contrato'
```

Memórias são persistidas por projeto, têm fonte e confiança e podem ser recuperadas por busca lexical. A camada de embeddings e retenção semântica será adicionada sem alterar esse contrato.

## Agendamentos e webhooks

Um schedule cria missões recorrentes. O intervalo é limitado a 31 dias e o claim é idempotente no store.

```bash
curl -sS -X POST http://localhost:11434/api/agent/v1/schedules \
  -H 'Content-Type: application/json' \
  -d '{"objective":"verificar o projeto","interval_seconds":3600,"webhook_secret_env":"DZ23_WEBHOOK_SECRET"}'
```

Para disparar por evento, configure a variável secreta no ambiente do processo e envie o header `X-Ollama-Agent-Secret`. O valor nunca é gravado no schedule.

```bash
export DZ23_WEBHOOK_SECRET='valor-fora-do-repositorio'
curl -sS -X POST http://localhost:11434/api/agent/v1/webhooks/SCHEDULE_ID \
  -H "X-Ollama-Agent-Secret: $DZ23_WEBHOOK_SECRET" \
  -H 'Content-Type: application/json' \
  -d '{"event":"push","ref":"main"}'
```

## Tools da primeira fatia

`workspace.list` lê entradas do workspace autorizado e limita a quantidade retornada. `workspace.read` lê no máximo 1 MiB e rejeita traversal. `workspace.write` exige approval, limita o payload, escreve com permissões restritas e cria um manifesto com tamanho e SHA-256. `terminal.exec` existe em modo conservador e só permite `pwd`, `ls` e `git status` com argumentos separados; shell interpolation e outros executáveis são bloqueados. `sandbox.exec` executa Python ou Node em user namespace, sem rede, com um bind apenas do workspace e limite de tempo/output; ainda exige approval.

Browser, computer use, MCP, conectores, mídia e código em sandbox dedicado ainda são adapters das fases seguintes. A API não finge que essas ferramentas estão disponíveis antes de receber implementação e testes reais.
