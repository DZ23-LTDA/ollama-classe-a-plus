# Inspeção visual — 2026-09-22

A primeira captura após a evolução funcional encontrou uma regressão: `class-a-plus-home.png` mostrava a tela global de erro `Something went wrong` com `Failed to fetch`, porque a rota raiz aguardava a API de settings mesmo sem backend. `index.tsx` foi corrigido para continuar em modo local-first quando a API estiver offline, preservando apenas o redirect de onboarding quando a resposta existir.

A captura de `class-a-plus-projects.png` confirmou que o menu lateral desktop já está aberto por padrão e visualmente organizado com Nova tarefa, Agente, Tarefas, Agendado, Habilidades, Plugins, Biblioteca, Projetos, Apps e providers, Workspace e Configurações. A tela de Projetos tinha formulário de criação e estrutura visual adequada, mas exibia apenas o erro literal `Failed to fetch` quando o backend não estava disponível; o fluxo agora continua mostrando a tela e registra estado de indisponibilidade, sem bloquear toda a aplicação.

A captura foi ajustada para aguardar `domcontentloaded` e um intervalo curto, pois `networkidle` não é apropriado para superfícies com polling agentic contínuo.

## Segunda inspeção

Após a correção, `class-a-plus-home.png` abre corretamente e mostra o shell desktop com menu lateral, título central “O que posso fazer por você?”, composer real, atalhos para slides/site/design/jogos e cards de recomendações. Sem o backend Ollama ligado, o composer permanece desabilitado e o ModelPicker mostra `Loading...`; isso é um estado de dependência do runtime, não uma simulação de execução.

`class-a-plus-plugins.png` mostra a tela funcional de Plugins com consulta ao catálogo agentic, blocos separados para Connectors allowlisted e MCP stdio, política de secrets/approvals/cross-tenant e estado vazio honesto quando não há configuração. A faixa `Failed to fetch` ainda pode aparecer quando a API não está rodando; a tela continua utilizável e não afirma que um connector/MCP existe.

## Terceira inspeção contra servidor local

Com Ollama em `127.0.0.1:3001`, a home abriu e o composer deixou de mostrar `Loading...`; sem modelos instalados, o picker mostra `Select a model`, que é o estado correto para instalar/selecionar um modelo.

A rota Plugins encontrou uma regressão real: erro global `Cannot read properties of null (reading 'length')`. A causa provável é o catálogo `/api/dz23/cli-catalog` ou algum recurso agentic retornando `null` em vez de array; o frontend precisa normalizar todas as respostas (`Array.isArray`) antes de calcular contagens. Esta falha deve ser corrigida antes de declarar Plugins funcional.

## Quarta inspeção contra servidor local

A tela Plugins deixou de quebrar após normalizar `null` para arrays; Connectors e MCP aparecem como estado vazio real, sem crash.

O Agentic Console ainda mostrava `Not Found` porque o frontend consulta `/api/agent/v1/metrics`, mas o servidor registrava somente `/metrics/prometheus`. A correção necessária é registrar também o endpoint JSON `/metrics`; sem isso a criação de missão fica visualmente bloqueada pelo erro global, mesmo com os outros contratos disponíveis.

## Quinta inspeção final

Depois do registro de `GET /api/agent/v1/metrics`, `class-a-plus-agentic.png` carrega sem `Not Found`; mostra Nova tarefa com selects de motor e projeto, métricas, orquestração multiagente e pesquisa profunda. A missão smoke anterior elevou o contador de criadas para 2, confirmando consulta ao runtime local.

`class-a-plus-settings.png` mostra o Agentic Control Center online, provider Ollama no catálogo, approvals server-side, workspace isolado, catálogo de connectors/MCP/skills vazio e GitHub Copilot detectado como CLI. A mensagem de configuração nativa é informativa e não bloqueia o painel agentic.


## Sexta inspeção — Company OS e Remote MCP

`class-a-plus-company.png` foi capturada com Chromium contra Vite e o servidor Ollama local em `127.0.0.1:3001`. A tela mostra o menu persistente, a empresa criada pelo smoke, status ativa, backlog aberto, KPI, ciclo ativo, budget, identidade, departamentos virtuais, ciclo e guardrails de segurança. A tela não afirma que CRM, anúncios, redes sociais, afiliados, dropshipping ou OAuth estejam conectados.

O smoke `scripts/smoke-company-os.sh` passou por criação, leitura, edição, roadmap, meta, backlog, ciclo ligado ao schedule, report, rejeição de gasto sem approval, gasto aprovado e pause/resume. O smoke Chromium `app/ui/app/scripts/smoke-shell.mjs` também passou por home, Projetos, Agendado, Plugins com catálogo MCP local/remoto, Skills, Empresa e Agentic Console.


## Sétima inspeção — capacidades operacionais e lifecycle

As capturas finais `class-a-plus-company.png` e `class-a-plus-plugins.png` foram produzidas em 2026-09-22 com Chromium contra Vite e o servidor Ollama local endurecido. A tela Company mostra Growth OS sandbox com métricas, campanha, programa de afiliados, produto e pedido persistidos; os textos informam que publicação, anúncio, parceiro e fulfillment exigem approval e connector autorizado. A tela Plugins mostra catálogo real vazio no ambiente sem manifests, política ativa de approvals, secrets fora da interface e cross-tenant rejeitado no servidor. A ausência de connectors e MCP nesta captura é o estado correto do ambiente de teste, não uma simulação de recursos conectados.
