# Entrega da fase 6 — endurecimento, multimídia e builder visual

## Implementado nesta continuação

A fase adiciona MFA TOTP com segredo cifrado em repouso, verificação server-side por request e redaction do ciphertext nas respostas; provisionamento inicial OIDC por userinfo HTTPS, criação idempotente de usuário/organização e renovação de credencial já cifrada; OCR local opcional via Tesseract com path containment, limites e manifesto de artifact; canvas visual declarativo com componentes, propriedades, filhos, posição/tamanho, preview HTML isolado, manifesto `visual.json`, versionamento e endpoint de atualização; exportação profissional mínima e verificável para PDF, DOCX e PPTX, além do ZIP existente; e a síntese comparativa dos padrões reimplementáveis dos harnesses avaliados.

## Endpoints adicionados ou ampliados

| Método | Endpoint | Função |
|---|---|---|
| `POST` | `/api/agent/v1/auth/mfa/enable` | Habilita TOTP usando segredo Base32 fornecido pelo operador autenticado. |
| `POST` | `/api/agent/v1/auth/mfa/disable` | Remove MFA da conta autenticada. |
| `POST` | `/api/agent/v1/media/ocr` | Executa Tesseract local quando instalado e gera artifact textual. |
| `POST` | `/api/agent/v1/builders/:id/visual` | Persiste árvore visual, regenera preview e incrementa a versão. |
| `POST` | `/api/agent/v1/builders/:id/export/pdf` | Gera PDF mínimo com conteúdo do projeto. |
| `POST` | `/api/agent/v1/builders/:id/export/docx` | Gera pacote DOCX OOXML mínimo. |
| `POST` | `/api/agent/v1/builders/:id/export/pptx` | Gera pacote PPTX OOXML mínimo. |

O fluxo OAuth existente agora aceita `OLLAMA_AGENT_OAUTH_<PROVIDER>_USERINFO_URL`. O provisionamento público de primeiro login só é permitido quando `OLLAMA_AGENT_AUTH_SSO_PUBLIC=true`; caso contrário o fluxo permanece vinculado a uma sessão autenticada. OIDC ainda requer configuração real do issuer, client, callback, TLS, escopos e política de domínio. SAML não é declarado concluído nesta fase; deve ser integrado por adapter licenciado e testado, sem implementação criptográfica caseira.

## Provas executadas

- `gofmt -w internal/agent/*.go server/agent_routes.go`
- `go test ./internal/agent -count=1` — suíte agentic passou após correção do `Root` do builder.
- `go test ./server ./cmd/launch ./internal/multillm -count=1` — deve ser reexecutado após o último guard de tenant e registrado no checkpoint somente após conclusão.
- Testes de MFA, exportadores, canvas visual e contratos existentes estão no pacote `internal/agent`.
- Teste distribuído `distributed_integration_test.go` permanece com build tag `integration` e exige PostgreSQL, Redis e collector reais; não é substituto do gate unitário.

## Limites honestos

A geração de PDF/DOCX/PPTX é um exportador base de contrato, não ainda uma suíte editorial equivalente a PowerPoint, Word ou Typst. O OCR depende de Tesseract instalado. OIDC userinfo não substitui discovery, validação completa de issuer/audience/nonce e política de domínio. SAML, push de produção com credenciais reais, testes físicos de companions e publicação em lojas/clouds ainda exigem ambientes e credenciais do proprietário. A síntese de harnesses orienta arquitetura, mas não permite copiar internals proprietários nem afirmar paridade de produto.
