# Security

The Ollama maintainer team takes security seriously and will actively work to resolve security issues.

## Reporting a vulnerability

> **Escolha o canal certo pelo componente afetado.** Este repositório é o fork
> **Ollama Classe A+** (DZ23), com uma superfície agentic própria sobre o Ollama.

### A) Vulnerabilidade no núcleo do Ollama upstream

Se a falha está no motor Ollama original (não na camada agentic/Classe A+),
reporte ao canal do upstream, **não** a este repositório: envie para
hello@ollama.com. Dê tempo hábil para investigação antes da divulgação pública.

### B) Vulnerabilidade específica do Ollama Classe A+ (este fork)

Se a falha está na camada Classe A+ — runtime agentic, connectors, MCP, deploy,
Company OS, HarnessRouter, companion, UI/mobile deste repositório — **não abra
issue pública** e **não** use hello@ollama.com. Reporte de forma privada por:

1. **GitHub Private Vulnerability Reporting** — aba **Security** deste
   repositório → **Report a vulnerability** (recomendado; cria um advisory
   privado com histórico e correção coordenada).
2. Se o botão acima não aparecer, o recurso ainda não foi habilitado pelo
   mantenedor (veja a nota abaixo); nesse caso, contate de forma privada o
   mantenedor do repositório público **DZ23-LTDA/ollama-classe-a-plus** pelo
   GitHub para combinar um canal privado antes de qualquer divulgação.

Inclua sempre: descrição, passos de reprodução, impacto avaliado, mitigações
possíveis, e a **referência do commit + arquivo afetado**. Nunca coloque
credenciais reais em relatórios, issues, PRs, screenshots, fixtures ou logs.

> **Nota de habilitação:** o *GitHub Private Vulnerability Reporting* deste fork
> foi **habilitado em 2026-09-24**. Use o item B.1 (aba **Security → Report a
> vulnerability**). Não foi criado nenhum e-mail de segurança específico do fork.

Please include the following details in your report:
- A description of the vulnerability
- Steps to reproduce the issue
- Your assessment of the potential impact
- Any possible mitigations

## Security best practices

While the maintainer team does its best to secure Ollama, users are encouraged to implement their own security best practices, such as:

- Regularly updating to the latest version of Ollama
- Securing access to hosted instances of Ollama
- Monitoring systems for unusual activity

## Contact

For any other questions or concerns related to security, please contact us at hello@ollama.com

## Ollama Classe A+

O runtime agentic pode executar ferramentas, acessar conectores, controlar companions, manipular arquivos e publicar builders. Em instalações públicas, mantenha `OLLAMA_AGENT_AUTH_REQUIRED=true`, use TLS/mTLS quando houver dispositivo remoto, configure PostgreSQL RLS por organização, mantenha tokens em um secrets manager e não habilite modos `DEV` ou `ALLOW_INSECURE` fora de loopback.

Vulnerabilidades envolvendo SSRF, path traversal, bypass de approval, fuga entre organizações, exposição de tokens, execução fora da sandbox ou publicação não autorizada devem ser tratadas como alta prioridade. Não coloque credenciais reais em issues, PRs, screenshots, fixtures ou logs. Para o código específico desta distribuição, encaminhe também a referência do commit e o arquivo afetado ao mantenedor do repositório público.
