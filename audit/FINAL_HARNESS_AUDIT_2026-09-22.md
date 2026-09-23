# Auditoria final do Harness — Ollama Classe A+

**Data:** 2026-09-22
**Escopo:** plugins e conectores; MCP stdio e Remote MCP; skills, tools e permissões; providers e sessões; UI desktop/mobile; Company OS e Growth OS; segurança, autenticação e infraestrutura; QA, CI, packaging e release.
**Natureza:** consolidação de auditorias somente leitura. **Nenhum código foi editado.**

> **Conclusão executiva:** o projeto possui um núcleo local expressivo, com adapters, testes unitários relevantes e documentação relativamente honesta sobre várias dependências externas. Entretanto, **não está pronto para ser declarado Classe A+ nem pronto para usuário final em produção**. Há blockers críticos de segurança, wiring, multi-tenant, protocolo, release e validação de jornada. A implementação comprovada é principalmente local e determinística; os adapters de terceiros não equivalem a integrações validadas.

## 1. Conclusão executiva

O checkout demonstra uma base funcional para execução local controlada. O runtime registra tools com risco e aprovação, há controles de autenticação, OAuth/PKCE, MFA, persistência atômica local, limites de resposta e vários testes focados. O registry de providers e os adapters de clientes externos existem. A UI desktop tem rotas e jornadas reais, e o companion mobile possui cache, fila offline e aprovação otimista. Esses elementos são evidência de **implementação parcial validada**, não de disponibilidade operacional ampla.

A fronteira de segurança ainda não é suficiente para exposição multiusuário. A autenticação agentic é opcional e desligada por padrão; quando isso ocorre, rotas mutáveis e superfícies operacionais podem ficar acessíveis sem identidade. A proteção de workspace é lexical e não impede symlink, hardlink ou TOCTOU. O sandbox usa namespaces de forma best-effort, sem quotas, seccomp ou garantia de disponibilidade e contenção. Remote MCP aceita destinos HTTPS sem uma política SSRF robusta. Skills podem ser marcadas como confiáveis pelo chamador sem atestado criptográfico, e os metadados de scopes não são aplicados como política de capability.

A frente MCP é um adapter parcial, não um fluxo MCP completo. Os testes focados cobrem helper process, allowlist, JSON/Bearer, SSE simples e rejeição de HTTP externo. Não foi localizado o wiring que transforme `OLLAMA_AGENT_MCP` e `OLLAMA_AGENT_REMOTE_MCP` em managers registrados no bootstrap. O Remote MCP também não implementa o conjunto completo de Streamable HTTP/OAuth, e não valida de forma robusta IPs privados, redirects, DNS rebinding ou correlação de IDs JSON-RPC.

A frente Company/Growth é um sandbox local persistente. Ela modela ciclos, campanhas, pedidos, estoque, aprovações e reports, mas não publica anúncios, envia mensagens, cobra, compra ou altera marketplaces. Algumas rotas mutáveis não demonstram `companyForRequest` antes da operação, o que requer bloqueio até que isolamento por organização seja provado por testes HTTP.

O baseline de release está vermelho. `go test ./...` falha em `cmd/TestFormerAgentEntryPointsAreRejected/agent`, pois o teste espera que o comando `agent` seja rejeitado enquanto a implementação o oferece. A UI desktop teve 195 de 199 testes passando, com 4 falhas em 2 arquivos. O typecheck mobile não executou porque `tsc` não está instalado e não há testes mobile inventariados. A cadeia de release também não demonstra gates obrigatórios para toda a suíte, smoke pós-empacotamento, assinatura, SBOM e provenance.

Assim, a classificação honesta é:

| Dimensão | Estado consolidado | Interpretação |
|---|---|---|
| **Adapter** | Presente em várias áreas | Há wrappers, registries e clientes para providers, conectores, MCP, CLI e UI. O adapter traduz contratos; não prova que o terceiro está disponível ou autorizado. |
| **Sandbox local** | Parcialmente funcional | Há fixtures, stores JSON, helper processes e testes locais. O sandbox local não deve ser confundido com isolamento forte, HA ou produção. |
| **Implementação local validada** | Parcial | O núcleo `internal/agent`, `internal/multillm` e vários testes focados passam em execuções registradas. Há conflitos entre configurações de build, portanto o checkout não é globalmente verde. |
| **Integração externa validada** | Não comprovada | Não houve smoke E2E confiável com providers, Composio, Remote MCP real, OAuth tenant-aware, contas sociais, dispositivos mobile ou runners multiplataforma. |
| **Release Classe A+** | Não atingido | Há achados críticos e gates internos vermelhos. |

Os números abaixo são **achados únicos consolidados**, com duplicatas entre frentes mescladas: **15 CRITICAL, 32 HIGH, 33 MEDIUM e 6 LOW**. A lista `FALHAS=[]` recebida não elimina as falhas de testes e blockers descritos nas auditorias por área; ela apenas indica que não foram fornecidas falhas adicionais fora desses resultados.

## 2. Matriz de cobertura e estado por área

| Área | O que está implementado | O que foi validado localmente | O que continua sendo apenas adapter ou sandbox | Integração real | Veredito |
|---|---|---|---|---|---|
| Plugins e conectores | `ConnectorManager`, OAuth organizacional, Connector HTTP, Remote MCP e adapters de launch | Testes de validação de IDs, URLs, métodos, limites, redaction, allowlists e configuração | Egress, credenciais, hosts configuráveis e contratos de terceiros | Não validada contra GitHub, Google, Slack, Discord, WhatsApp ou Composio | **Parcial; não expor sem política de egress e tenant** |
| MCP stdio/Remote MCP | Helper process, JSON-RPC linear, POST/SSE simplificado, Bearer opcional | `go test -race ./internal/agent -run 'MCP' -count=1` passou; fixtures locais passaram | Não há loader/wiring comprovado; não há sandbox forte; Remote MCP não é protocolo completo | Não houve servidor MCP externo, OAuth PKCE, sessão ou pairing real | **Blocker de integração e segurança** |
| Skills, tools e permissões | Registry de tools, risk/approval, workspace, terminal allowlisted, sandbox best-effort, manifest de skills | Testes de approval, workspace, sandbox Python, browser, connector e MCP | Skills são catálogo/manifest; `Trusted` não é atestado; scopes não são enforcement | Não provado com skill executável assinada ou política multi-tenant | **Núcleo útil, fronteira de confiança incompleta** |
| Providers e sessões | Registry/proxy OpenAI-compatible, OmniRoute/HarnessRouter adapters, queue, mission, schedule e browser session | `go test ./internal/multillm ./server ./internal/agent -count=1` passou em uma execução; auth e proxy têm regressões | JSON local, adapters CLI e fixtures não demonstram HA, contratos upstream ou capabilities reais | Sem provider real, HarnessRouter, OmniRoute, Composio ou browser físico | **Local/sandbox; não pronto para operação distribuída** |
| UI desktop | Rotas Agentic/Company/Projects/Library/Scheduled/Skills/Plugins/Settings e adapter HTTP central | 195 de 199 testes passaram; build foi executado, mas a suíte terminou falhando | Dados interceptados e sessão local; token ainda depende de storage no renderer | Sem login enterprise, deployment, provider e teste de release | **Jornada não verde** |
| Mobile | Companion de missão, approval, timeline, cache, fila offline, `If-Match` e polling | Não houve typecheck: `tsc: not found`; não há testes unitários/E2E no app | Expo/EAS, SecureStore/AsyncStorage e polling são implementação de companion | Sem dispositivo físico, push real, lojas ou E2E Android/iOS | **Não validado** |
| Company OS | Store JSON, lifecycle, roadmap, ciclos, budget, pausa/resume, approvals básicos | `CGO_ENABLED=0 go test ./internal/agent -count=1` passou; smoke local cobre happy path | Growth é sandbox; effects externos são inexistentes; JSON não é HA | Sem Meta/X/YouTube/WhatsApp/TikTok/Shopify/marketplaces | **Sandbox local, não sistema operacional externo** |
| Segurança e infraestrutura | AuthStore, SHA-256 de tokens, revogação, OAuth PKCE, MFA, AES-GCM, RBAC básico e middleware | Testes de auth, OAuth, MFA, recovery, approval e isolamento básico | Segurança de produção depende de flags, TLS/proxy, store distribuído e host | Sem OIDC/SAML/infra distribuída/mTLS/observabilidade real | **Não fail-closed por padrão** |
| QA, CI, packaging e release | 352 arquivos `_test.go`, matrizes de release, artefatos e checksums | Muitos pacotes passam; release unitária tem cobertura relevante | Build Docker, toolchains, downloads e jobs GitHub são adapters | Sem GPU/OS runners/registry/releases reais e sem assinatura verificável | **Release bloqueado** |

## 3. Achados CRITICAL

Os achados críticos abaixo devem ser encerrados antes de exposição externa, ativação multi-tenant ou publicação de release. A ordem é de risco e dependência, não de conveniência.

### C-01 — Autenticação agentic não é fail-closed

`OLLAMA_AGENT_AUTH_REQUIRED` é `false` por padrão. Em um listener exposto ou atrás de reverse proxy, missões, projetos, companies, builders, traces, métricas e operações mutáveis podem ser acessíveis sem identidade. No mesmo modo, `missionByID` não tem uma barreira de organização defensiva. A configuração segura depende de disciplina operacional não imposta pelo produto.

### C-02 — Workspace e artifacts podem escapar por symlink/TOCTOU

`safeWorkspacePath` e `resolveWorkspace` fazem contenção lexical com `Abs/Join/isWithin`, mas não resolvem symlinks nem revalidam o caminho no momento de uso. Um symlink dentro do workspace pode direcionar leitura ou escrita para fora da raiz. O teste existente cobre traversal lexical, não symlink, hardlink, junction ou corrida entre validação e uso.

### C-03 — O sandbox local não é isolamento forte comprovado

`sandbox.exec` usa `unshare`, namespaces e bind mount, mas não há seccomp/AppArmor, cgroups, limites de CPU/memória/PIDs, quota de disco, filesystem read-only ou política de syscalls. Código arbitrário pode consumir recursos e modificar o workspace montado. A capacidade também depende do kernel e de permissões do host. O produto não pode apresentar isso como sandbox de segurança forte sem um executor isolado equivalente.

### C-04 — Remote MCP tem SSRF crítico

O adapter aceita HTTPS externo sem resolver e verificar todos os endereços contra loopback, RFC1918, link-local, multicast e metadata endpoints. O `http.Client` padrão segue redirects, permitindo que o destino efetivo mude depois da validação inicial. Há risco de atingir rede privada ou metadata e de reenviar headers/Bearer a um destino não autorizado.

### C-05 — O carregamento documentado de MCP não foi encontrado

Não foi localizado wiring executável que leia `OLLAMA_AGENT_MCP` ou `OLLAMA_AGENT_REMOTE_MCP`, valide os manifestos, registre os servidores e os torne disponíveis no bootstrap de `Runtime`. A documentação promete a capacidade, mas os presets podem resultar em `/mcp` vazio e tools indisponíveis.

### C-06 — Remote MCP não é um fluxo completo de Streamable HTTP/OAuth

O adapter remoto injeta um Bearer opcional, mas não implementa OAuth 2.0/PKCE, refresh, sessão, pareamento, revogação, `initialize`, versão de protocolo, `MCP-Session-Id`, resumption, reconnect ou encerramento de sessão. O comportamento atual é um POST/SSE simplificado e não deve ser rotulado como integração MCP completa.

### C-07 — Resposta JSON-RPC não é correlacionada ao request

Não há validação do ID da resposta contra o ID solicitado. Resposta atrasada, de outro request ou de um servidor malformado pode ser aceita. O stdio também mantém mutex durante a chamada e lê uma única linha, sem roteamento por ID ou multiplexação.

### C-08 — Trust de skill é caller-controlled e não há enforcement de capability

`SkillManifest.Trusted` pode ser sobrescrito pelo booleano recebido em `LoadSkills`. Não há assinatura, hash, origem, owner, atestado ou validação de scopes/tools. Além disso, scopes e trust são metadados: o Runtime aplica principalmente `RequiresApproval` e risco, não uma política de capability por missão, skill, projeto e organização.

### C-09 — Rotas mutáveis de Company/Growth não provam isolamento por organização

As operações de approve/launch/pause de campanha, approve de programa, conversão, approve de pedido e fulfill não chamam de forma demonstrada `companyForRequest` antes da mutação. Sem proteção global equivalente e testes HTTP cross-tenant, existe risco de bypass de RBAC ou acesso entre organizações.

### C-10 — Approval não está vinculado a ator, organização e política

A aprovação é essencialmente estado/boolean ou uma decisão que valida `missionID`, `approvalID` e status. Não há prova de ator, tenant, motivo obrigatório, política aplicada, expiração, nonce, segregação de funções ou proteção contra replay. Isso não é suficiente para efeitos destrutivos ou externos em ambiente multiusuário.

### C-11 — Growth OS não executa os efeitos que a jornada sugere

A implementação local modela campanhas, conversões, pedidos, fulfillment e reports. Ela não publica anúncios ou posts, envia mensagens, cobra, compra, altera marketplaces, chama transportadoras ou gateways. Declarar a jornada como “concluída” para usuário final seria uma falsa equivalência entre simulação local e integração operacional.

### C-12 — Baseline Go completo está vermelho

`go test ./...` falha em `cmd/TestFormerAgentEntryPointsAreRejected/agent`: o teste espera que `ollama agent` seja rejeitado, mas o comando existe e retorna sucesso. O contrato precisa ser reconciliado antes de qualquer claim de release verde.

### C-13 — A cadeia de release não demonstra gates obrigatórios suficientes

O material auditado não comprova que `go test ./...`, `go vet`, lint, `git diff --check`, smoke do binário empacotado e validação de artefatos sejam pré-condições obrigatórias antes da publicação. Um pipeline pode produzir ou publicar apesar do baseline vermelho.

### C-14 — Artefatos e imagens não têm assinatura, SBOM e provenance verificáveis

Há checksums SHA-256, mas não foram encontradas assinaturas Cosign/Sigstore, SBOM, attestation de provenance ou verificação de identidade do publisher. Checksum sem assinatura não estabelece quem produziu o artefato nem impede substituição do manifesto.

### C-15 — Jornada de release de UI/mobile não está validada

A suíte desktop terminou com 18 arquivos passando e 2 falhando, totalizando 195 testes aprovados e 4 falhos. A divergência inclui `pl-36` esperado contra `pl-6` renderizado. O mobile nem chegou ao typecheck por falta de `tsc` e não tem testes registrados. A experiência final não pode ser declarada pronta.

## 4. Achados HIGH

Os achados HIGH devem entrar no backlog imediato. Alguns são condições de fechamento dos críticos; outros impedem a operação confiável mesmo depois de a exposição básica estar protegida.

1. **MCP stdio executa comando configurado com permissões do usuário.** Não há allowlist de executáveis/paths, sandbox de filesystem/rede, limite de memória ou controle de recursos por subprocesso.
2. **Allowlist MCP opcional é fail-open.** Quando `AllowedMethods` está vazio, qualquer método JSON-RPC pode ser chamado. Registro com configuração vazia deve falhar.
3. **Connectors permitem `base_url` configurável sem política de egress forte.** Faltam pinning de host, proteção contra DNS rebinding, controle de redirects e garantia de que token não atravesse a fronteira de origem autorizada.
4. **`ConnectorManager.Call` não exige `organizationID`.** Uma API exportada sem contexto de tenant pode virar bypass futuro de isolamento, mesmo que a tool use `CallForOrganization`.
5. **Respostas HTTP 4xx/5xx de connector não têm semântica uniforme de erro.** Falta estrutura de erro, política de retry, redaction de corpo/headers e garantia de não propagação de conteúdo sensível para eventos/artifacts.
6. **Transporte stdio é linear e frágil.** Mantém mutex durante I/O, lê exatamente uma linha e não suporta notificações, respostas fora de ordem, progresso, multiplexação ou restart robusto.
7. **Limites de payload são incompletos.** Há limite de resposta e body, mas não há limite explícito e uniforme para request, params, profundidade, número de eventos SSE ou alocações antes de processar.
8. **Não há trilha de auditoria específica por chamada MCP.** Faltam `server_id`, método, approval, tenant, latência, status e redaction por tentativa e resultado.
9. **Registro/listagem/chamada MCP não demonstram enforcement tenant-aware.** A camada HTTP/RBAC externa não deve ser a única barreira defensiva.
10. **Capabilities MCP não são declaradas nem validadas.** A política deveria permitir apenas ferramentas e métodos explicitamente autorizados por servidor, missão, skill e organização.
11. **Browser e desktop agregam muitos efeitos em uma única aprovação.** Falta autorização granular por ação, teste de fuga de arquivos, takeover e processos.
12. **Não há E2E contra provider real.** Streaming, retries/fallback, tool calls, cancelamento, timeout, 401/429/5xx e autenticação real não foram exercitados contra OpenAI-compatible, OmniRoute, HarnessRouter, xAI ou Composio.
13. **Catálogo de providers não distingue disponibilidade real.** O registry não prova que endpoint, credencial, model ID e capabilities declaradas existem no upstream.
14. **Sessões, missões, schedules e fila usam JSON local sem HA.** Não há locking distribuído, leases, recuperação após crash, idempotência de efeitos externos ou concorrência entre workers.
15. **CLI adapter permite execução local configurada pelo operador.** Faltam allowlist de executáveis, `cwd/env` controlados, timeout obrigatório, limites de saída e cancelamento de process group.
16. **Fila mobile grava bearer token em AsyncStorage.** Um item offline pode conservar token em backup/exportação ou após logout; o flush deveria recuperar credencial do SecureStore e invalidar itens antigos.
17. **Mobile não possui cobertura automatizada.** Não há testes para approval, logout, troca de servidor, offline flush, conflito 409, notificações ou Android/iOS reais; o typecheck está bloqueado por dependência ausente.
18. **Polling e fila mobile podem gerar tempestade ou duplicação.** O polling fixo de 3 segundos não tem backoff, abort ou mutex; a fila não tem idempotency key, limite de tentativas, dead-letter ou tratamento de erros permanentes.
19. **`pushRegistered` é global demais.** Troca de servidor, organização ou usuário pode impedir novo registro; o estado deve incluir contexto e ser limpo no logout.
20. **Desktop guarda bearer e organização em localStorage.** Um XSS no renderer teria acesso ao token; preferir bridge/secure storage, CSP explícita, validação de origem e testes de expiração/401.
21. **Company/Growth não têm idempotência, reserva atômica ou rollback suficiente.** Fulfillment repetido pode baixar estoque novamente; pedidos concorrentes podem fazer oversell; `AddCycle` pode deixar ciclo sem schedule após falha intermediária.
22. **A suíte release é condicionada por tags, hardware e recursos.** Casos `integration && release` podem ser skipped por VRAM, arquitetura, tempo ou ausência de daemon, reduzindo a força do gate.
23. **Downloads de toolchains e instaladores não são pinados por hash ou assinatura.** O Dockerfile baixa Go derivado de `go.mod`, e workflows instalam CUDA/ROCm/Vulkan/cuDNN e Windows por URLs externas.
24. **Docker usa tags mutáveis e não promove por digest.** `ollama/ollama`, versões e `-rocm` não estabelecem imutabilidade; não há assinatura ou validação pós-push completa.
25. **Publicação destrutiva não tem dry-run ou verificação forte do commit.** O script localiza release por nome, apaga assets antigos e altera tag com `contents: write`.
26. **A lista de artefatos obrigatórios é incompleta.** O gate não exige explicitamente todos os bundles Darwin/Linux esperados e pode divergir da matriz de produção.
27. **Disponibilidade do sandbox não é um gate explícito.** Hosts sem user namespace, mount, Node ou Python podem desabilitar a capacidade, e hosts permissivos não têm prova de contenção contra OOM/fork bomb.
28. **OAuth aceita redirect URI absoluta com host sem allowlist por provider/tenant.** Isso requer proteção contra open redirect, CSRF e mistura de state/provider.
29. **MFA e recovery não têm rate limit ou lockout.** Um token válido com brute force online continua sendo uma possibilidade operacional.
30. **`/connect` é liberado por bypass genérico do middleware.** A segurança passa a depender totalmente de `deviceConnect`; faltam testes negativos de token, Origin, revogação, TLS e mTLS no handshake.
31. **AuthStore local tem locking apenas em memória.** Múltiplos processos podem perder atualizações ou aceitar decisões concorrentes; o modo PostgreSQL/Redis é opcional e não foi validado com RLS.
32. **Adapters externos carecem de controles operacionais completos.** Ainda faltam webhooks assinadas, deduplicação, DLP, quotas, rate limits, rollback e reconciliação por provider.

## 5. Achados MEDIUM

1. A canonicalização de paths não cobre de forma ampla query strings, percent-encoding, barras duplicadas, Unicode e roteamento divergente.
2. Faltam testes de concorrência, cancelamento durante I/O, respostas fragmentadas/grandes, SSE sem `data`, JSON-RPC batch, redirects e headers sensíveis.
3. Faltam testes de crash, timeout, EOF, JSON inválido, ID incorreto, restart e concorrência no stdio.
4. Faltam testes remotos de redirect, DNS/private IP, timeout, cancelamento, body excedente, sessão, auth negativa e SSE malformado.
5. `StopAll` pode aguardar processos sob lock sem deadline; um processo que ignore encerramento pode bloquear limpeza.
6. Não há estado local explícito que diferencie registrado, autenticado, conectado e upstream-ready.
7. `openCodeStatePath` possui TODO e hardcodes XDG Linux/macOS, com risco no Windows.
8. `go test ./app/tools` não compila no Linux porque os arquivos têm build constraints `windows || darwin`; falta CI correspondente.
9. A documentação não traz matriz de permissões, retenção/redaction, rotação/revogação de tokens, quotas ou rollback por preset.
10. `findConnectorOperation` usa `HasPrefix` literal; `/users` pode autorizar `/users-privileged` sem correspondência por segmento.
11. Identificadores de MCP, nomes de métodos e profundidade de params não têm validação específica além dos limites gerais.
12. Logs e eventos podem carregar erros, respostas e headers sensíveis sem redaction uniforme.
13. Registry pode sobrescrever silenciosamente tool com mesmo nome; `Descriptors` não tem ordem determinística.
14. Adapters de launch são majoritariamente unitários e não validam instalação, upgrade, versões atuais e contratos de terceiros.
15. Não há matriz automatizada de compatibilidade para `/v1/chat/completions`, `/v1/responses`, tool calls, reasoning e streaming por provider.
16. Não há testes de rotação/expiração de credenciais, falha de upstream, retry budget, circuit breaker e observabilidade sem prompt/segredo.
17. Não há browser session contra navegador ou desktop real; takeover depende de Playwright e superfície conectada.
18. UI não tem cobertura suficiente de viewport, breakpoints, teclado virtual, safe areas, tablets, contraste, screen reader, foco e teclado.
19. Mobile aceita URL HTTP arbitrária fora de uma política clara de localhost/LAN confiável; tokens podem trafegar em claro.
20. `agentFetch` não define timeout/abort próprio e assume JSON, sem contrato para malformed, timeout, 401/403 e CORS.
21. Timeline mobile não tem paginação/limite visível nem tratamento explícito de datas inválidas e payloads parciais.
22. Company não impõe unicidade de SKU, limites de campos, consistência de moeda ou proteção de overflow em totais/receita.
23. Reports não têm reconciliação, período explícito, paginação ou trilha de cálculo.
24. JSON local não oferece locking entre processos, migração/schema versioning, recuperação de corrupção ou transações de produção.
25. Handlers Company podem retornar objeto completo sem política dedicada de campos e redaction.
26. Smokes Company/Growth cobrem happy path e approval básico, não autorização, concorrência, repetição, restart ou falhas parciais.
27. Não há pipeline mínimo de PR claramente demonstrado com `go test`, vet, lint e validação rápida antes de jobs caros.
28. A documentação não define critérios, modelos, VRAM, duração, skips e cobertura por backend para `integration && release`.
29. Há pacotes sem testes em áreas relevantes de onboarding, extract-examples, readline, generator, batch e arquiteturas.
30. Checksum usa `xargs sha256sum` sem `-print0`, com risco para nomes contendo whitespace; o manifesto não é allowlist assinada.
31. Não há smoke pós-build de cada binário/imagem, instalação/desinstalação, layout de libs ou `--version` em ambiente isolado.
32. StepID é usado para formar arquivos `.agent-sandbox` antes de validação estrita; valores com separadores ou paths absolutos devem ser rejeitados.
33. Métricas, traces e endpoints operacionais ficam públicos quando auth está desligada; logs e pid files têm permissões amplas e não há redaction uniforme.

## 6. Achados LOW

1. Exemplos e documentação ainda usam presets mutáveis como `npx --yes @wonderwhy-er/desktop-commander@latest`, o que é inadequado para produção, embora seja um risco que deve ser tratado junto do hardening de supply chain.
2. A documentação mistura linguagem de “adapter concluído” com linguagem que pode ser lida como integração pronta; a separação entre catálogo, autenticado e conectado deve ser visual e operacional.
3. Alguns estados de UI e catálogo não distinguem claramente “configurado”, “credenciado”, “upstream indisponível” e “não testado”.
4. Mockups e notas de captura podem ser confundidos com tela implementada se não houver badge de estado em todos os canais de produto.
5. Quotas, retenção e lifecycle de credenciais não estão apresentados como defaults operacionais no produto.
6. A ordenação não determinística de alguns descritores e a ausência de versionamento explícito de adapters dificultam diagnósticos, embora não sejam o bloqueio principal.

## 7. Quick wins internos

Os quick wins abaixo não exigem credenciais de terceiros e devem ser executados antes de qualquer nova integração externa:

1. **Mudar o default de exposição para fail-closed.** Exigir auth em listeners não-loopback; manter modo local sem auth apenas em loopback e bloquear mutações/integrações quando não houver identidade.
2. **Tornar `AllowedMethods` obrigatório e não vazio.** Falhar no registro e no load de configuração vazia para MCP local e remoto.
3. **Corrigir o contrato do comando `agent`.** Decidir entre remover/renomear o subcomando ou atualizar teste e documentação; repetir `go test ./...`.
4. **Corrigir os 4 testes de UI e o segundo arquivo falho.** Fazer o layout esperado corresponder deliberadamente ao design e transformar build/test em gate.
5. **Implementar o loader de `OLLAMA_AGENT_MCP` e `OLLAMA_AGENT_REMOTE_MCP`.** Validar JSON estrito, limites, erros e registro no bootstrap; adicionar teste E2E local com `Runtime` real.
6. **Bloquear SSRF e redirects no Remote MCP.** Rejeitar loopback, RFC1918, link-local, metadata, multicast e IPv6 especiais após resolução; revalidar cada conexão; não reenviar Bearer a outro origin.
7. **Validar correlação de ID JSON-RPC e sanear o ciclo de vida stdio.** Separar reader/writer, suportar cancelamento e descartar notificações sem confundir respostas.
8. **Sanear `StepID` e implementar containment resistente a symlink.** Usar ID gerado pelo servidor, `EvalSymlinks`/`openat`/`NOFOLLOW` quando disponível e testes de TOCTOU.
9. **Introduzir `CapabilityPolicy` central e aprovação vinculada.** Exigir actor, tenant, motivo, expiração, nonce e policy; negar capability não declarada.
10. **Remover bearer da fila mobile e reduzir exposição desktop.** Guardar apenas referência no AsyncStorage, limpar no logout e preparar bridge/secure storage para o renderer.
11. **Fixar dependências executadas.** Remover `@latest`/`--yes` de produção, registrar versão e digest, e verificar hash de downloads de toolchains.
12. **Adicionar gate mínimo de release.** `git diff --check`, `go test ./...`, `go vet ./...`, lint, smoke `--version`, manifesto de payloads e falha explícita quando qualquer etapa é skipped.
13. **Corrigir `runtime.GOOS` no OpenCode e adicionar teste Windows.** Colocar runners Windows/macOS no gate para os caminhos platform-gated.
14. **Adicionar redirect allowlist e rate limiting.** Restringir callback OAuth por provider/tenant e impor backoff/lockout a MFA/recovery.
15. **Documentar estados de integração.** Exibir separadamente adapter instalado, sandbox local testado, credencial presente, upstream healthy e E2E aprovado.

## 8. Backlog ordenado de remediação

### P0 — antes de exposição externa ou release

**Objetivo:** retirar riscos de execução não autenticada, escape de arquivos, SSRF e release vermelho.

- Fechar auth por default para qualquer listener não-loopback e proteger `/connect` no próprio handshake.
- Corrigir containment de workspace/artifacts e sanitização de StepID.
- Definir se o sandbox é limitado a ambiente local confiável ou substituí-lo por worker/container/VM com seccomp, cgroups, filesystem temporário read-only e limites.
- Implementar política SSRF/redirect para Remote MCP e connectors.
- Tornar allowlists MCP obrigatórias e validar correlação de IDs.
- Resolver o loader de manifestos MCP e testar o bootstrap real.
- Vincular approval a ator, tenant, policy, motivo, expiração e replay protection.
- Corrigir `go test ./...`, os testes de UI e obter typecheck mobile em workspace com dependências instaladas.
- Bloquear publicação enquanto os gates obrigatórios não forem executados e aprovados.

### P1 — hardening do núcleo local

**Objetivo:** tornar a execução local previsível e auditável.

- Criar `CapabilityPolicy` com deny-by-default, scopes por missão/skill/projeto/organização/tool e conflitos explícitos no registry.
- Restringir MCP stdio por manifesto confiável, executável, cwd, ambiente mínimo e limites de CPU/memória/PIDs/output.
- Implementar lifecycle MCP com reader dedicado, cancelamento, restart, framing, notificações e payload limits.
- Adicionar auditoria redacted para tools, MCP, approval, connectors, browser e desktop.
- Usar storage seguro para bearer no desktop e remover permissões amplas de logs/pid.
- Adicionar testes de symlink, TOCTOU, resource exhaustion, malformed JSON/SSE, concorrência e cross-tenant.
- Tornar Company/Growth state machine, idempotente e atomicamente consistente; reservar estoque e testar concorrência.

### P2 — adapters e contratos de providers

**Objetivo:** provar que cada adapter faz o que anuncia sem confundir fixture com upstream.

- Criar suíte parametrizada de contrato para chat, Responses, SSE, tools, reasoning, timeout, cancel, 401/429/5xx e redirects.
- Separar catálogo, credenciado e upstream-ready; expor health e last-error sem segredos.
- Publicar matriz de compatibilidade por provider/model/path/capabilities.
- Definir retry budget, circuit breaker, rotação de credenciais e redaction de observabilidade.
- Testar CLI adapters com process group, `cwd/env` restritos, allowlist de executável e argument injection.
- Implementar e testar OAuth/PKCE, refresh, sessão, revogação e Streamable HTTP MCP conforme contrato aplicável.

### P3 — mobile, UI e acessibilidade

**Objetivo:** fechar a jornada de usuário final sem chamar mockup de produto pronto.

- Criar workspace mobile reprodutível com `npm ci`, typecheck, lint e testes.
- Adicionar testes de criação, refresh, approval, rejeição, logout, cache, fila offline, 409, troca de servidor e push.
- Implementar backoff, `AbortController`, mutex, idempotency keys, TTL, dead-letter e status por item na fila.
- Exigir HTTPS fora de localhost/LAN explicitamente confiável e exibir aviso de transporte inseguro.
- Adicionar E2E de viewport, teclado, foco, labels, contraste, screen reader e reduced motion.
- Validar desktop com 401/expiração, CSP, origem, redaction e secure storage.

### P4 — multi-tenant e operação distribuída

**Objetivo:** substituir garantias locais por garantias operacionais reais.

- Tornar PostgreSQL/RLS, locking transacional, leases e Redis obrigatórios para modo multiusuário.
- Persistir estado de execução, idempotência e recovery após crash.
- Testar authorization por tenant em todas as rotas de mission, schedule, session, browser, Company, artifacts e builders.
- Adicionar observabilidade sem payloads sensíveis, retenção definida, quotas e runbooks de revogação/rollback.
- Validar mTLS, reverse proxy, TLS, OIDC/SAML e webhooks assinadas em ambiente de staging.

### P5 — integração externa e release confiável

**Objetivo:** demonstrar disponibilidade de terceiros e supply chain verificável.

- Executar smoke real, com credenciais de teste, para cada provider aprovado e documentar data, versão, tenant, quotas e resultado.
- Validar Remote MCP/Desktop Commander com pairing, OAuth, sessão, revogação e dispositivo controlado.
- Validar Social/Growth com sandboxes, webhooks, deduplicação, returns/refunds, rate limits e reconciliação.
- Fixar downloads por digest/assinatura; gerar SBOM, provenance e assinatura de binários, manifests e imagens.
- Publicar por digest imutável, exigir environment protection e manter rollback testado.
- Executar matriz Linux/macOS/Windows, CPU/GPU/backends e instalação/desinstalação em runners reais.

## 9. Blockers externos

Estes itens não podem ser resolvidos apenas no checkout local, mas precisam de dono, ambiente e evidência antes de uma declaração de pronto:

- **Credenciais e contas de providers:** xAI, Cerebras, Mistral, OpenAI-compatible, OmniRoute, HarnessRouter, Composio e demais serviços requerem chaves, quotas, model IDs e permissões reais.
- **Remote MCP e Desktop Commander:** dependem de endpoint disponível, OAuth 2.0/PKCE, conta, máquina pareada, sessão Bearer válida, agente em execução e controles de revogação.
- **Integrações de produtividade e canais:** GitHub, Google Workspace, Slack, Discord e WhatsApp Cloud exigem tokens, escopos, tenant, quotas e rede válidos.
- **Growth e commerce:** Meta/Instagram, X/Twitter, YouTube, TikTok Shop, Shopify e marketplaces exigem app review, sandboxes, webhooks assinadas, pagamentos, fulfillment e regras regionais.
- **HA e produção:** PostgreSQL, RLS, Redis, OpenTelemetry, TLS/mTLS, reverse proxy, gestão de segredos e múltiplos workers não estão presentes no sandbox local.
- **Mobile e distribuição:** Expo/EAS, Android SDK, iOS/Xcode, dispositivos físicos, push provider, certificados, App Store Connect e credenciais de assinatura são necessários. O `ascAppId` real e credenciais EAS ainda bloqueiam publicação.
- **Build e backends:** CUDA, ROCm, Vulkan, MLX/Metal, cuDNN, Windows/macOS runners, Docker multi-platform e hardware GPU não estão disponíveis para validação local completa.
- **Supply chain:** registries, mirrors, downloads de Go/SDKs/instaladores e terceiros npm precisam de disponibilidade, hashes/assinaturas e política de promoção.
- **Terceiros e contratos:** schemas, versões, quotas, disponibilidade, TLS/PKI e comportamento de redirects estão fora do controle do repositório e devem ser registrados como evidência externa, não presumidos.

## 10. Gates de pronto

Nenhum gate deve ser considerado satisfeito por documentação ou fixture isolada. Cada gate precisa de log, versão do checkout, ambiente, credencial/tenant de teste quando aplicável e resultado reproduzível.

### Gate G0 — baseline determinístico

- `go test ./...` passa sem atualizar testes para esconder regressão.
- `go vet ./...`, lint e `git diff --check` passam.
- Build reproduzível do binário e smoke `--version`/startup passam.
- UI desktop build e testes passam sem arquivos falhos.
- Mobile typecheck, lint e suíte mínima passam em workspace com `npm ci`.

### Gate G1 — segurança de exposição

- Auth é obrigatória fora de loopback e todas as rotas mutáveis exigem tenant/actor.
- `/connect` valida token, Origin, revogação, TLS/mTLS conforme deployment.
- Workspace/artifacts não escapam por symlink, hardlink, junction ou TOCTOU.
- Sandbox possui limite de recursos e falha de forma segura quando isolamento não está disponível.
- Logs, métricas, traces, headers, bodies, cookies e tokens são redacted.

### Gate G2 — MCP e connectors

- Loader de manifestos está coberto por teste de bootstrap real.
- Allowlists não vazias, capabilities declaradas e IDs correlacionados são obrigatórios.
- Remote MCP implementa o contrato de transporte escolhido, com sessão, reconnect e auth documentados.
- SSRF, DNS rebinding, redirect, private IP, link-local, metadata, IPv6 e header forwarding têm testes negativos.
- Cada chamada registra approval, tenant, server ID, método, status, latência e redaction.

### Gate G3 — autorização e estado

- Approval carrega actor, organização, policy, motivo, timestamp, expiração e nonce.
- Capability policy é deny-by-default e aplicada no Runtime, não apenas em catálogo/planner.
- Company/Growth têm state machine, idempotency key, reserva de estoque, rollback/compensação e testes cross-tenant.
- PostgreSQL/RLS/Redis e leases são obrigatórios no modo multiusuário, com teste de crash e concorrência.

### Gate G4 — providers e integrações

- Cada provider aprovado tem contrato automatizado e smoke externo separado do fixture local.
- Há evidência de chat, Responses, streaming, tool calls, reasoning, retry/fallback, cancelamento, 401/429/5xx e rotação de credencial.
- O sistema distingue catalogado, credenciado, conectado, healthy e E2E aprovado.
- Webhooks são assinadas e deduplicadas; quotas, DLP, redaction e rollback têm runbook.

### Gate G5 — UI e mobile

- Jornadas de criação, approval, rejeição, polling, offline, logout, troca de organização/servidor e conflito 409 passam.
- Não há bearer em AsyncStorage; desktop usa secure storage/bridge com CSP e limpeza após logout.
- Há testes de acessibilidade, viewport, foco, teclado, safe area, orientação e reduced motion.
- Push, fallback polling e permissões são testados em dispositivo ou em harness equivalente declarado.

### Gate G6 — release e supply chain

- Todos os artefatos esperados por plataforma/arquitetura/backend são produzidos e verificados contra manifesto.
- Downloads têm hash ou assinatura verificada; imagens são promovidas por digest imutável.
- Binários, imagens e manifests têm assinatura, SBOM e provenance verificáveis.
- Instalação, upgrade, desinstalação, smoke pós-build e rollback passam em runners reais.
- Publicação destrutiva exige aprovação, valida tag/commit e possui procedimento de recuperação.

## 11. Definição honesta de pronto para usuário final

O produto estará **pronto para usuário final em produção** somente quando todas as condições seguintes forem verdadeiras:

1. O usuário final, organização e operador são autenticados por padrão quando o serviço não está estritamente limitado a loopback.
2. Nenhuma tool, skill, connector, MCP server, browser session, builder ou Company route consegue escapar da organização, workspace ou capability concedida.
3. Toda ação externa ou destrutiva exige approval vinculada a ator, tenant, policy, motivo, expiração e evento de auditoria.
4. O sandbox oferece contenção declarada e testada; se a plataforma não oferecer isolamento forte, a capacidade é desabilitada e o usuário recebe erro explícito, não uma falsa promessa de segurança.
5. MCP, connectors e providers mostram o estado real de configuração, autenticação, conectividade e saúde. Adapter instalado não aparece como integração pronta.
6. Cada integração anunciada tem contrato e smoke E2E com credenciais de teste, e o runbook documenta quotas, revogação, rotação, erros, rollback e limites.
7. Sessions, missions, schedules, approvals e efeitos externos são idempotentes, recuperáveis após crash e seguros com múltiplos workers.
8. Desktop e mobile passam build, typecheck, testes, acessibilidade e jornadas críticas em ambientes representativos; offline queue e push não expõem ou reutilizam tokens indevidamente.
9. `go test ./...`, UI, mobile e gates de release estão verdes. Testes skipped são explicitamente justificados e não contam como aprovação.
10. O artefato entregue é reproduzível ou atestado, assinado, acompanhado de SBOM, provenance, checksums e matriz de compatibilidade.
11. O usuário recebe mensagens honestas sobre capacidades locais, integrações não conectadas, credenciais ausentes, limites de segurança e ações que exigem operador.

**Estado atual:** não pronto para usuário final em produção; no máximo, **preview local controlado para desenvolvedores e auditores**, com integrações externas tratadas como não validadas. “Classe A+” só deve ser usada depois de fechar os achados P0 e comprovar G0–G6.

## 12. Evidências e arquivos centrais

A consolidação foi baseada nos resultados fornecidos e nos seguintes grupos de evidência do checkout:

- Runtime, tools e segurança: `internal/agent/runtime.go`, `internal/agent/tools.go`, `internal/agent/types.go`, `internal/agent/context.go`, `internal/agent/auth.go`, `internal/agent/store.go`, `server/agent_routes.go`, `server/auth.go`.
- MCP e connectors: `internal/agent/mcp.go`, `internal/agent/mcp_remote.go`, `internal/agent/connectors.go`, respectivos testes e `examples/dz23-desktop-commander-*.json`.
- Providers e sessões: `internal/multillm/registry.go`, `internal/multillm/proxy.go`, `server/dz23_multi_provider_test.go`, `internal/agent/queue.go`, `internal/agent/browser.go`.
- Company/Growth: `internal/agent/company.go`, `internal/agent/company_growth.go`, `server/company_routes.go`, `server/company_growth_routes.go` e scripts `smoke-company-*.sh`.
- UI/mobile: `app/ui/app/src/lib/agenticClient.ts`, `app/ui/app/src/components/layout/layout.test.tsx`, `apps/mobile-agentic/App.tsx`, `apps/mobile-agentic/package.json`, `.github/workflows/dz23-multi-provider.yaml`.
- QA/release: `.github/workflows/release.yaml`, `Dockerfile`, `scripts/build_docker.sh`, `scripts/push_docker.sh`, `integration/reg_release_test.go`, `cmd/cmd_test.go`, `CMakePresets.json`.
- Documentação: `docs/agentic/ARCHITECTURE.md`, `docs/agentic/API.md`, `docs/agentic/INTEGRATIONS.md`, `docs/dz23-multi-provider.md`, `docs/development.md`, `SECURITY.md` e `README.md`.

## Referências

[1]: file:///home/ubuntu/work/ollama-dz23-work "Checkout do projeto Ollama auditado"
[2]: file:///home/ubuntu/work/ollama-dz23-work/audit "Diretório de auditorias e evidências do projeto"
[3]: file:///home/ubuntu/work/ollama-dz23-work/docs/agentic/ARCHITECTURE.md "Arquitetura agentic documentada no checkout"
[4]: file:///home/ubuntu/work/ollama-dz23-work/.github/workflows/release.yaml "Workflow de release do checkout"
[5]: file:///home/ubuntu/work/ollama-dz23-work/SECURITY.md "Política de segurança do checkout"


## Addendum de remediação incremental — após a auditoria ampliada — 2026-09-22

Este documento preserva os achados originais como histórico. As slices publicadas posteriormente mitigaram partes relevantes sem transformar o produto em produção-ready.

| Achado original | Estado após as slices publicadas | Limite que permanece |
|---|---|---|
| C-01 autenticação não fail-closed | **Mitigado parcialmente**: bind não-loopback força auth e há regressões de dev-token remoto | TLS/reverse proxy, OIDC/SAML real e operação distribuída ainda dependem do ambiente |
| C-02 workspace/artifacts por symlink/TOCTOU | **Mitigado parcialmente**: tools, Builder preview/export e `BuildArtifactManifest` rejeitam symlink/realpath externo | TOCTOU e prova multi-plataforma/distribuída ainda não estão encerrados |
| C-03 sandbox forte | **Parcial**: process group, cancellation, cwd privado, stderr DLP e limites best-effort | seccomp/AppArmor/cgroups/quotas e isolamento forte permanecem abertos; não chamar de sandbox forte |
| C-04/C-06 Remote MCP/SSRF | **Mitigado parcialmente**: proxy nil, redirect same-origin e IP efetivo são validados | OAuth/session/revocation, Streamable HTTP completo e todos os egresses ainda precisam de prova |
| C-08 skills/plugins | **Mitigado parcialmente**: lifecycle autenticado tem ownership por organização | capability enforcement, assinatura/trust e registro server-owned ainda são pendentes |
| C-09 Company/Growth tenant state | **Mitigado em slices**: Company create, Builder, orchestration/traces/devices, plugins e jobs têm negativos cross-tenant | todas as mutações e prova Postgres RLS/Redis real ainda não foram exercitadas |
| C-10 approvals | **Mitigado em missão e Company OS**: owner/admin, nonce, CAS e ledger de gasto | políticas ABAC/segregação por ação e efeitos externos reais ainda não estão concluídos |
| C-12 baseline Go | **Verde localmente**: `CGO_ENABLED=1 go test ./...`, vet e build passaram nas slices | runner GitHub e Docker distribuído continuam evidência externa |
| C-13/C-14 release supply chain | **Mitigado parcialmente**: quality gate, artefato não vazio, checksum e attestation condicional | signing, SBOM final, provenance executada e publicação real não foram comprovados |
| C-15 UI/mobile | **Verde em gates locais**: Vitest/build/typecheck mobile | dispositivos físicos, distribuição e jornadas reais continuam `NOT_RUN` |

**Classificação vigente:** preview/local RC em hardening. Este addendum não revoga os blockers externos nem autoriza alegações de credenciais, contas, devices, deploys, marketplace, app review, signing ou release concluídos.


## Addendum — Browser Operator upstream CI — 2026-09-22

O run upstream `35750983274` falhou em `go test`/`go test -race` porque o workflow não instalava o módulo Python `playwright`, importado por `browser_helper.py`; o log confirmou `ModuleNotFoundError`. O workflow agentic do fork já instalava essa dependência e permaneceu verde. O patch adiciona Playwright `1.63.0` e Chromium aos jobs upstream relevantes, com instalação de dependências OS somente no Linux, e torna o launcher Go compatível com `python3` ou `python` no `PATH`. A matriz nativa Linux/Windows foi tornada manual e opt-in por depender de runners compatíveis. A correção trata CI/portabilidade; não prova Browser/desktop completo, sandbox forte ou release production-ready. A confirmação remota do novo head permanece pendente.


## Addendum — workflow upstream preso e dependência Playwright — 2026-09-22

O run `35755046119` foi cancelado depois de confirmar que a matriz `linux`/`windows` dependia de runners customizados não disponíveis no PR público. O mesmo run mostrou a causa independente dos dois jobs Ubuntu: a pinagem `playwright==1.53.2` não era publicada no índice acessível ao runner. A remediação usa `playwright==1.63.0`, corrige o grupo de concorrência e restringe a matriz nativa a execução manual opt-in. A validação local passou, mas a prova remota no novo head e a homologação multiplataforma permanecem abertas.


## Addendum — validação remota final do CI upstream — 2026-09-22

A remediação do Browser Operator e do lint condicional por plataforma foi confirmada no head `de0e86772e96372789c10d924eb5738f8808821b`. O workflow upstream `test` passou no run `35769597404`, incluindo testes em Linux, macOS e Windows, race em Linux e macOS, patches e `go_mod_tidy`. Os workflows `class-a-plus-integrity` (`35769597628`), `dz23-agentic-quality` (`35769597363`) e `dz23-multi-provider` (`35769597578`) também passaram. O PR reporta 21 checks bem-sucedidos, 3 skipped, 0 failing e 0 pending.

A causa foi corrigida sem esconder falhas: o Browser Operator agora consegue localizar o Chromium gerenciado pelo Playwright; o lint foi saneado no código compartilhado e nos arquivos condicionais Darwin/Windows; e o erro de credencial protegido foi declarado somente no build não-Windows que o utiliza. Os gates locais também passaram com `golangci-lint v2.13.2`, Go test normal/race, vet, build e compilação cruzada Windows dos pacotes afetados.

Esta evidência fecha o gate do **caminho normal de CI**. Ela não fecha a matriz GPU/nativa, que continua manual e depende de runners compatíveis, nem os blockers de produção listados neste documento. Permanecem abertos sandbox/process isolation forte, auth/session/CSRF/IdP distribuído, OAuth lifecycle/revocation, egress/DLP residual, providers/deploy/media reais, dispositivos físicos, signing/provenance, stores e app review. A classificação vigente continua **preview/local RC em hardening; não production-ready**.


## Addendum — hardening Company/session/terminal e CI do head 935fb273 — 2026-09-22

As slices posteriores mitigaram lacunas internas específicas sem apagar os achados originais. A UI deixou de tratar `403` como expiração de sessão; `POST /api/agent/v1/auth/logout` revoga bearer tokens no servidor; o planner usa o contrato efetivo do cliente Ollama; spend, conversão de afiliado e métrica social possuem ledger de idempotência por digest/fingerprint, com replay seguro e conflito explícito; e o terminal allowlisted valida flags e paths antes da execução. Essas correções melhoram retry, logout e autorização operacional, mas não transformam a sessão em um sistema IdP distribuído nem o processo best-effort em sandbox forte.

A prova remota do head `935fb273` é: integrity `35784211426` (também push `35784205466`), agentic quality `35784211429` (também push `35784205382`), multi-provider `35784211481` e upstream `test` `35784211483`; todos passaram. O upstream executou testes em Linux, macOS e Windows, race em Linux/macOS, patches e `go_mod_tidy`. A matriz GPU/nativa não foi executada e permanece manual/opt-in.

Os blockers permanecem: seccomp/cgroups/quotas e isolamento forte de processos; auth/session/CSRF/IdP distribuído; OAuth lifecycle/revocation e refresh reais; egress/DLP residual; providers, deploy/media e marketplaces reais; testes físicos desktop/mobile; signing/provenance; stores e app review. A classificação vigente continua **preview/local RC em hardening; não production-ready**.


## Addendum — sandbox strict, Origin policy, mobile tenant scope e release metadata — 2026-09-22

As slices `0459532a`, `e931fcfc`, `a12a15e4`, `d1058209` e `a7bc82f5` mitigaram lacunas internas específicas. O novo strict sandbox Linux requer cgroup v2 delegado, falha fechado quando indisponível e encadeia namespaces, `no-new-privs`, seccomp amd64/arm64, limites e cgroup kill no timeout. Isso é uma fronteira implementada e testada localmente, não prova de AppArmor/SELinux ou de enforcement em host de produção. A política de Origin reduz CSRF em mutations autenticadas; o mobile confirma tenant via auth/session e namespacifica estado offline/push; o release workflow acrescenta SBOM CycloneDX, metadata, checksum manifest e verificação antes do upload.

A evidência remota disponível no head `a7bc82f5` é integrity push `35793674459` PASS e integrity PR `35793679458` PASS. Os demais workflows do PR ainda estavam em execução/fila no momento do checkpoint. O workflow de release não foi executado porque exige tag e ambiente de release; attestation é condicional, e signing/provenance efetiva não foi comprovada. O pacote mobile reportou 18 vulnerabilidades de produção no npm audit (11 moderate, 7 high), que precisam de triagem antes de classificar distribuição como pronta.

A classificação vigente permanece **preview/local RC em hardening; não production-ready**. Permanecem abertos host sandbox real e isolamento forte multi-plataforma, auth/session/CSRF/IdP distribuído, OAuth egress/revocation com providers reais, integrações/deploy/media/marketplaces, push e dispositivos físicos, dependências mobile, instaladores assinados, provenance/rollback efetivos, stores/app review e homologação externa.


## Addendum — follow-up de dependências mobile — 2026-09-22

O audit de produção inicialmente reportou 18 vulnerabilidades transitivas. A análise de ranges mostrou que o fix sugerido exigia major do Expo/React Native; foram preferidos overrides compatíveis para `image-size@2.0.4`, `postcss@8.5.28` e `uuid@11.1.1`, com lock resolvido. O estado atual passa `npm audit --omit=dev` com zero vulnerabilidades, typecheck e Expo web export. Isso não comprova compatibilidade física Android/iOS nem elimina a necessidade de uma futura migração major controlada.


## Addendum — CI normal verde no head e6e0632b — 2026-09-22

O head `e6e0632ba5ff0495ed4b061221c90696478b3d67` passou upstream `test` (`35794443450`), `class-a-plus-integrity` (`35794443307`), `dz23-multi-provider` (`35794443300`) e `dz23-agentic-quality` (`35794443501`). O upstream passou Linux, macOS, Windows, race Linux/macOS, patches e go_mod_tidy; o workflow agentic passou Go/server, Browser Operator, PostgreSQL RLS + Redis DLQ + OTLP, Web/Mobile e SBOM. O PR consolidou 21 successful, 3 skipped, 0 failing e 0 pending.

O gate de integrity agora protege explicitamente sandbox strict, isolamento mobile e release SBOM/checksum; os gates Node usam `npm ci`. Isso é evidência do caminho normal de CI, não um release assinado nem homologação dos skips. A matriz GPU/nativa continua manual/opt-in e permanecem abertos sandbox de host, IdP/OAuth/providers/deploy/media reais, dispositivos, push remoto, signing/provenance efetiva, rollback, stores/app review e produção.


## Addendum — V5 provider, connectors e CI normal — 2026-09-22

As mudanças V5 mitigaram lacunas internas específicas sem revogar os achados originais. O planner agora exige um provider/modelo resolvido e falha fechado em caso de ausência ou erro. O estado durável foi separado do workspace, o Remote MCP usa IPs aprovados após DNS e o catálogo de connectors informa apenas readiness seguro por tenant. A UI deixou de sugerir aliases Claude/Codex/OmniRoute como conectados quando o runtime não publica um modelo correspondente.

A sequência de CI também foi corrigida com diagnóstico observável. O head `6d8b8710` encontrou um helper planner morto; `93a0115f` o removeu. O head seguinte revelou uma colisão transitória do cache npm global no Windows; `0f95b6a1` isolou o cache por runner e desabilitou cancelamento em cascata nas matrizes test/race. No head final passaram upstream `test` (`35804229207`), `class-a-plus-integrity` (`35804229189`), `dz23-multi-provider` (`35804229204`) e `dz23-agentic-quality` (`35804229301`).

O resultado fecha o caminho normal de CI desta slice, não o produto inteiro. Permanecem abertos sandbox de host homologado, auth/IdP distribuído, OAuth e providers externos, Composio e connectors de contas reais, Woovi/OpenPix e fiscal/NF-e, social commerce/marketplaces, deploy/media, dispositivos físicos, push remoto, signing/provenance, rollback, stores e app review. A classificação vigente continua **preview/local RC em hardening; não production-ready**.


## Addendum pós-V5 — correção macOS e registro durável de connectors — 2026-09-22

A conclusão do head `0da6be80` corrige a leitura anterior de CI pendente. O workflow upstream `test` run `35808941810` terminou com sucesso, incluindo `test (macos-latest)` job `107015911723`, além de normal Ubuntu `107015911678`, normal Windows `107015911689`, race macOS `107015880499` e race Ubuntu `107015880534`. Os checks fork integrity `35808941807`, agentic quality `35808941824` e multi-provider `35808941746` também passaram. A causa macOS foi um falso positivo da validação final de workspace diante do alias `/var` → `/private/var`; `os.Lstat` passou a verificar somente a entrada final, enquanto containment e rejeição de componentes descendentes foram mantidos. O teste de configuração passou a usar o diretório efetivo da plataforma.

Após esse head, o commit `411335ba` implementou o primeiro lifecycle server-owned para connectors. A ausência de `OLLAMA_AGENT_CONNECTORS` agora cria um manager com manifest durável em `OLLAMA_AGENT_STORE/connectors.json`; mutations são serializadas com arquivo temporário, sync, rename e rollback, usando modo `0600`. O conteúdo é limitado a configurações não secretas. O endpoint `POST /api/agent/v1/connectors` usa `DisallowUnknownFields`, aceita apenas nomes de env e IDs OAuth, exige owner/admin no modo autenticado, força o tenant da sessão e impede que um ID existente de outro tenant seja sobrescrito. A UI Plugins expõe o cadastro sem coletar token.

A evidência local do novo slice inclui testes de round-trip do manifest, permissões, redaction, rollback, colisão cross-tenant, membro não administrador, organização conflitante e raw-secret rejection, além de testes Go focados, integrity e build UI. O novo SHA iniciou seus checks remotos, mas a auditoria não declara esses checks como finais até a conclusão pública. O manifest estático apontado por `OLLAMA_AGENT_CONNECTORS` continua sendo uma fonte de bootstrap explícita e não deve ser descrito como persistência do cadastro da UI.

O achado de lifecycle foi reduzido somente para connectors. MCP stdio, Remote MCP e skills ainda precisam de registro/persistência equivalente, e o sandbox de processo continua best-effort fora do modo strict Linux provisionado pelo operador. OAuth, contas Composio/Google/GitHub/Woovi/OpenPix/NF-e, social commerce, marketplaces, deploy, hardware, signing, stores e app review continuam não validados. O estado correto segue **FIXING / preview-local RC em hardening**; não é final nem production-ready.


## Addendum — MCP/Remote MCP/skills durable lifecycle — 2026-09-22

O commit `96fd7edd` adicionou `POST /api/agent/v1/mcp`, `POST /api/agent/v1/remote-mcp` e `POST /api/agent/v1/skills`. Em auth mode, owner/admin é obrigatório e `organization_id` é derivado do contexto autenticado. O manager recusa colisão de ID cross-tenant. O runtime padrão grava MCP em `OLLAMA_AGENT_STORE/mcp.json`, Remote MCP em `OLLAMA_AGENT_STORE/remote-mcp.json` e skills em `OLLAMA_AGENT_STORE/context/skills`; arquivos são produzidos atomicamente e com modo `0600`.

A validação stdio exige executável absoluto regular e workspace seguro. A validação Remote MCP preserva HTTPS fora de loopback, SSRF/private-address rejection, DNS resolution/pinning, same-origin redirects, métodos allowlisted e tokens/headers somente por nome de env. O registro de skills força `trusted=false`, deriva enabled no servidor e não aceita campos de autoridade do cliente. A UI Plugins possui cadastro operacional para os três tipos sem copiar segredos.

A cobertura executada foi: testes de round-trip/restart, permissões, rollback, tenant collision, trust fail-closed, authorization e raw-secret/unknown-field rejection, normal e race; runner local completo PASS em integrity/YAML/Go/UI/mobile/audit/diff. O SHA novo foi publicado e seus checks remotos ainda estavam pendentes no momento da nota. O modo `OLLAMA_AGENT_MCP`/`OLLAMA_AGENT_REMOTE_MCP` segue bootstrap estático e não apresenta mutation de UI como alteração durável do arquivo indicado.

Riscos não reduzidos nesta slice: OAuth real e refresh/revoke com contas, Desktop Commander ou Composio em upstream, providers externos, social/commerce/fiscal, attestation de skill, isolamento físico de processos/dispositivos, IdP distribuído, homologação em hardware, signing, stores, deploy e app review. Estado: **FIXING / preview-local RC em hardening**, não final e não production-ready.


## Addendum P1 — parsing estrito de manifests MCP — 2026-09-22

A revisão do novo storage encontrou uma ambiguidade de decoder: os loaders de MCP e Remote MCP paravam após o primeiro array JSON. O commit `89b4203e` adicionou uma segunda leitura e exige `io.EOF`; conteúdo trailing, incluindo um segundo objeto, falha fechado. O caso `[] {}` tem regressão dedicada nos testes de persistência.

O teste normal/race dos managers e o gate completo Go passaram em integrity, `go test ./...`, vet, build e diff. O comportamento de bootstrap estático não foi relaxado e nenhuma credencial é envolvida. O achado foi classificado como hardening P1 corrigido; blockers externos e o estado preview/local RC permanecem.
