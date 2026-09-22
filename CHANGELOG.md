# Changelog — Ollama Classe A+

Este arquivo registra as entregas públicas da distribuição `DZ23-LTDA/ollama-classe-a-plus`. O projeto mantém a atribuição e a licença do Ollama upstream; os recursos agentic específicos estão descritos com seus limites no [guia Classe A+](docs/CLASS_A_PLUS_GUIDE.md).

## [0.1.0] — Preview público

### Incluído

- Runtime agentic em Go com missões persistentes, planejamento, execução, observação, recovery, artifacts e approvals.
- Ferramentas com sandbox, contenção de workspace, políticas de SSRF/path e trilhas de auditoria.
- Browser Operator via Playwright, MCP stdio, memória, projetos, skills e scheduler/webhooks.
- UI Agentic Console para missões, métricas, multiagente, pesquisa e approvals.
- Base de companion desktop para Linux, macOS e Windows e cliente Expo mobile com cache/outbox offline.
- Conectores allowlisted, providers Claude/Codex, autenticação, RBAC, OAuth/OIDC/SAML/MFA, colaboração e device pairing.
- Filas, DLQ/replay, PostgreSQL/RLS, Redis, OTLP, WebSocket/mTLS e exportadores/builders.
- DeploymentManager com providers Vercel, Netlify e generic, approval obrigatório e controles de segurança de publicação.
- Documentação pública em português, arquitetura, API, integrações, roadmap, política de segurança e screenshots reais.

### Estado honesto do preview

Adapters, contratos, testes e documentação não significam que contas externas, credenciais, instaladores assinados, publicação em lojas, modelos multimídia ou testes físicos estejam disponíveis neste ambiente. Deploy real, SSO contra um IdP real, GPU/modelos locais e distribuição de companions permanecem validações dependentes de ambiente. Os mockups em `docs/images/mockups/` são conceitos; as limitações das screenshots reais estão em `docs/images/screens/SCREEN_CAPTURE_NOTES.md`.

### Verificação

Os gates comprovados desta revisão incluem testes focados do runtime/server, build Go, build/typecheck da UI e typecheck mobile em suas fases correspondentes. O workflow público [`dz23-agentic-quality`](.github/workflows/dz23-agentic-quality.yaml) executa os gates de qualidade e a integração distribuída no GitHub Actions.

[0.1.0]: https://github.com/DZ23-LTDA/ollama-classe-a-plus/releases/tag/v0.1.0
