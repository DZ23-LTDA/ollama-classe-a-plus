# Integrações externas e MCP

## Conectores HTTP

O runtime aceita um arquivo JSON apontado por `OLLAMA_AGENT_CONNECTORS`. O arquivo contém apenas metadados, URLs HTTPS, operações permitidas e o nome da variável de ambiente que contém o token. O valor do token nunca entra no repositório, missão, memória, evento ou resposta de capabilities.

```bash
export OLLAMA_AGENT_CONNECTORS=/etc/ollama-dz23/agent-connectors.json
export DZ23_GITHUB_TOKEN='...'
export DZ23_GOOGLE_TOKEN='...'
export DZ23_SLACK_TOKEN='...'
export DZ23_DISCORD_TOKEN='...'
export DZ23_WHATSAPP_TOKEN='...'
```

O exemplo [agent-connectors.json](../../examples/agent-connectors.json) cobre GitHub, Google Workspace, Slack, Discord e WhatsApp Cloud. Cada chamada exige uma operação declarada, método permitido e prefixo de caminho permitido. Redirects são desativados, o timeout é limitado e a resposta é limitada a 2 MiB. Operações de escrita são tools de efeito externo e exigem approval do runtime.

O fluxo de autenticação OAuth, renovação de refresh token e consentimento por usuário/organização deve ser adicionado à superfície de configuração antes de uso multiusuário. Tokens pessoais podem ser usados localmente apenas para desenvolvimento.

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
