# Backlog de paridade — Ollama Classe A+ vs. Manus, Codex e Claude

Levantado em 2026-09-25 testando o app instalado no Windows, página por página.
Cada item diz o estado real hoje e o que falta. "Feito" só é marcado com teste.

## Feito nesta rodada (PR #38)

- Páginas agentic (Agente, Tarefas, Agendado, Empresa, Biblioteca, Projetos)
  abriam em erro: o servidor da UI não fazia proxy de `/api/agent/*`.
- Tela de erro com Voltar / Início / Recarregar.
- App iniciava o `ollama.exe` do PATH (instalação antiga) em vez do embutido.
- Atualizador baixava o instalador oficial do Ollama e substituiria esta
  distribuição.
- Página **Conectores** com busca e abas, a partir do catálogo real.
- **Pesquisar** (menu e tecla `/`) em conversas e páginas.
- **Chat anônimo**: não salvo em disco nem no histórico.
- Anexo de **ZIP** (arquivos de texto internos, com limites) e anexos mais rápidos.
- Seletores legíveis no tema escuro; provedores sem chave marcados "(sem chave)".

## Lacunas priorizadas

| # | Capacidade | Manus / Codex / Claude | Estado aqui | Próximo passo |
|---|---|---|---|---|
| 1 | Agente no chat principal (ferramentas de arquivo, shell, código) | Os três executam ferramentas direto na conversa | Chat inicial só tem busca web; ferramentas ficam no Agente (Mission Console) | Unificar: permitir que o chat use o runtime agentic com approvals |
| 2 | Acesso ao agente com "Expose to the network" ligado | n/a | Exige login de workspace local; não há tela para criar a primeira conta | Fluxo de criação do primeiro usuário owner no Workspace |
| 3 | OAuth de conectores (Gmail, Google Workspace, GitHub, Slack, Notion...) | Um clique | Catálogo + registro manual com credencial do operador | Fluxo OAuth por conector com cofre de tokens |
| 4 | Computador/navegador controlado pelo agente | Manus e Claude operam browser e desktop | Browser Operator existe no runtime; sem UI de acompanhamento ao vivo | Painel de sessão do navegador com screenshots e takeover |
| 5 | Edição de código em repositório com diff/PR | Codex e Claude Code | Builder/artifacts existem; sem revisão de diff na UI | Visualizador de diff, aplicar/rejeitar, abrir PR |
| 6 | Artifacts ao vivo (sites, slides, apps) com preview | Claude Artifacts, Manus | Artifacts listados na Biblioteca sem preview | Preview em sandbox iframe com CSP |
| 7 | Memória persistente entre conversas | Claude/Manus | Memória por projeto no runtime; sem UI de gestão | Tela de memórias com editar/esquecer |
| 8 | Configurações em seções (Conta, Personalização, Conectores, Integrações, Dados) | Manus | Uma página longa | Reorganizar Configurações em navegação lateral |
| 9 | Integrações de mensagem (Slack, Telegram, WhatsApp) para enviar tarefas | Manus | Contratos no catálogo; sem bot | Bot Telegram primeiro (mais simples), depois Slack |
| 10 | Uso e faturamento por provedor | Manus/Claude | Custos no roteador multi-provider; sem tela | Painel de uso por provedor/modelo |
| 11 | Anexos grandes (vídeo, PDF grande, planilhas) | Os três | Limite de 10 MB por arquivo, sem xlsx/pptx | Upload por caminho local em streaming + extratores |
| 12 | Agendamento a partir do chat ("todo dia às 9h...") | Manus/Claude | Página Agendado exige formulário | Interpretar agenda em linguagem natural |

## Regras que continuam valendo

- Segredos nunca aparecem na UI nem em logs.
- Ações externas (enviar, publicar, pagar) exigem approval.
- Nada é mostrado como conectado sem registro real.
