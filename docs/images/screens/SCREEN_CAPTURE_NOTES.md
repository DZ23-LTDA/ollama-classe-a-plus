# Notas das capturas

`agentic-console.png` é uma captura real da rota `/agentic` da UI React atual, renderizada com Chromium e dados demonstrativos interceptados localmente. Ela mostra o Mission Console, criação de missão, métricas, orquestração multiagente, pesquisa profunda, timeline e approval pendente.

`settings.png` é a captura histórica da rota `/settings` anterior. Ela aparece praticamente vazia além do cabeçalho Settings e não deve ser apresentada como uma tela de configurações completa.

`class-a-plus-settings.png` é uma captura real da rota `/settings` após a primeira evolução do shell Classe A+. Ela mostra o **Agentic Control Center**, catálogo de providers quando disponível, estado sanitizado de approvals/workspace/auth e o aviso honesto de que a configuração nativa depende do backend Ollama. Ela não prova que credenciais, providers externos ou deploys estão configurados.

`class-a-plus-projects.png`, `class-a-plus-library.png`, `class-a-plus-scheduled.png`, `class-a-plus-skills.png`, `class-a-plus-plugins.png` e `class-a-plus-tasks.png` são capturas reais das novas rotas do shell, com o menu lateral Classe A+ aberto por padrão. Nesta rodada, essas superfícies têm estados vazios e ações de entrada; o conteúdo persistido e o CRUD completo ainda dependem da conexão dos contratos da API agentic. O script `app/ui/app/scripts/smoke-shell.mjs` verifica menu, títulos, links e ações primárias/empty-state.

As capturas usam dados demonstrativos e não representam uma execução contra um ambiente de produção. O README deve identificá-las como screenshots da implementação atual com dados demo.
