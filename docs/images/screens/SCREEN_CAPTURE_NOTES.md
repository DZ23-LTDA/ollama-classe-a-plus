# Notas das capturas

`agentic-console.png` é uma captura real da rota `/agentic` da UI React atual, renderizada com Chromium e dados demonstrativos interceptados localmente. Ela mostra o Mission Console, criação de missão, métricas, orquestração multiagente, pesquisa profunda, timeline e approval pendente.

`settings.png` é a captura histórica da rota `/settings` anterior. Ela aparece praticamente vazia além do cabeçalho Settings e não deve ser apresentada como uma tela de configurações completa.

`class-a-plus-home.png`, `class-a-plus-agentic.png` e `class-a-plus-settings.png` são capturas reais do shell, Agentic Console e Settings. A captura live do Agentic Console mostra seleção de motor/projeto, missão, métricas, orquestração e pesquisa; a Settings mostra o **Agentic Control Center**, catálogo de providers quando disponível e estado sanitizado de approvals/workspace/auth. Elas não provam que credenciais, providers externos, modelos ou deploys estejam configurados.

`class-a-plus-projects.png`, `class-a-plus-library.png`, `class-a-plus-scheduled.png`, `class-a-plus-skills.png`, `class-a-plus-plugins.png`, `class-a-plus-tasks.png` e `class-a-plus-company.png` são capturas reais das rotas do shell, com o menu lateral Classe A+ aberto por padrão. Contra o servidor local, Projetos e Agendado foram criados/listados/excluídos por CRUD real; Plugins e MCP exibem catálogo vazio sem crash quando o backend retorna `null`; Skills, Tasks e Biblioteca consultam contratos reais; Company OS mostra uma empresa criada pelo smoke, sete departamentos, um KPI, backlog, ciclo, budget e o painel Growth OS com campanha, programa, produto e pedido sandbox. Os scripts `app/ui/app/scripts/smoke-shell.mjs`, `scripts/smoke-company-os.sh` e `scripts/smoke-company-growth.sh` cobrem essas jornadas por API.

As capturas desta rodada foram produzidas contra Vite e um servidor Ollama local em `127.0.0.1:3001`, sem provider externo e sem modelo instalado. Elas demonstram a implementação e os estados locais, não uma execução de produção, deployment, login enterprise, dispositivo físico ou harness remoto.
