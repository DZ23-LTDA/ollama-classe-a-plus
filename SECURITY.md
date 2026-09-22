# Security

The Ollama maintainer team takes security seriously and will actively work to resolve security issues.

## Reporting a vulnerability

If you discover a security vulnerability, please do not open a public issue. Instead, please report it by emailing hello@ollama.com. We ask that you give us sufficient time to investigate and address the vulnerability before disclosing it publicly.

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
