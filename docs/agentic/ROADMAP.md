# Roadmap executável — Ollama DZ23 Agentic Platform

## Fase 0 — Base, contratos e gates

Consolidar os contratos versionados, checkpoint, build, testes, threat model, política de secrets e observabilidade. Esta fase inclui a ponte Claude/Codex já implementada e o primeiro pacote de documentação agentic.

## Fase 1 — Núcleo vertical de missões

Implementar o store, estados, eventos, planner, executor, observer, recovery, aprovação e artifact manifest. O fluxo comprovado será: criar missão, gerar plano, executar leitura autorizada, persistir resultado, produzir manifesto e consultar eventos.

## Fase 2 — Ferramentas seguras

Adicionar filesystem por workspace, terminal com executável e argumentos separados, execução de código em sandbox, limites de CPU/memória/tempo/rede, logs redacted e kill seguro. A execução arbitrária permanece bloqueada até a policy fornecer scope e aprovação.

## Fase 3 — Projetos, memória, skills e MCP

Criar projetos com permissões e contexto, memória episódica e semântica com retenção configurável, importação de arquivos, skill manifests assinados ou confiáveis e lifecycle de MCP servers. Cada fonte de contexto terá origem e nível de confiança.

## Fase 4 — Jobs e automações

Adicionar fila persistente, scheduler, retries, idempotency keys, webhooks verificados, dead-letter queue e replay. Integrações externas serão adapters com secrets server-side, scopes mínimos e confirmação para efeitos sensíveis.

## Fase 5 — Browser e computer use

Integrar browser isolado com perfis, downloads, uploads, navegação, screenshots, ações e takeover humano. Integrar Desktop companion para tela, mouse, teclado, clipboard e processos usando capability grants. Nenhuma dessas capacidades será simulada por texto.

## Fase 6 — Artefatos multimídia e builders

Adicionar documentos, slides, planilhas, gráficos, dashboards, imagens, áudio, voz, transcrição, vídeo, sites, aplicativos e jogos. Cada domínio deve possuir renderer, preview, export, manifest, checksum e smoke test.

## Fase 7 — Superfícies de produto

Construir API pública versionada, CLI agentic, Web/Desktop com timeline de missão, diff, approvals, logs, artifacts e terminal controlado. Construir aplicativo Mobile para chat, inbox de missões, approvals, notifications, artifacts e controle de projetos.

## Fase 8 — Integrações e colaboração

Adicionar GitHub, Google Workspace, e-mail, Slack, Discord, WhatsApp e outros connectors por adapters. Implementar organizações, usuários, papéis, RBAC/ABAC, auditoria, compartilhamento, comentários, presença e colaboração em tempo real.

## Gates por fase

Cada fase exige testes unitários, contratos, autorização negativa, integração real quando o adapter existir, segurança, observabilidade e documentação. Uma feature fica `PARTIAL` enquanto seu adapter não tiver execução real e smoke reproduzível.

## Ordem de investimento

A prioridade é segurança e recuperação, depois execução real, depois conectores e superfícies. A interface não será usada para mascarar lacunas de runtime. O próximo incremento executável é a Fase 1.
