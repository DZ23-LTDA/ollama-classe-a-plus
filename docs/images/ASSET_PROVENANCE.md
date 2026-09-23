# Proveniência dos assets visuais

Este diretório reúne imagens de origens diferentes porque a documentação do fork preserva parte da documentação de integrações herdada do Ollama. **Nem todo arquivo sob `docs/images/` é uma tela do Ollama Classe A+.** A classificação abaixo é a referência para README, revisões e futuras remoções.

## Assets do produto Classe A+

As capturas verificadas da interface do fork ficam em [`screens/`](screens/):

- `class-a-plus-*.png`: capturas observadas da UI web local-first, recapturadas com Chromium contra o Vite e o servidor local;
- `settings.png`: alias compatível da captura de Settings;
- `agentic-console.png`: captura histórica da superfície Agentic Console mantida para compatibilidade documental.

Essas imagens demonstram o estado observado no ambiente de captura. Elas não comprovam, sozinhas, credenciais externas, contas OAuth, dispositivos físicos, deploy, distribuição em lojas, homologação ou prontidão de produção. A data, o comando de captura e os limites estão em [`screens/SCREEN_CAPTURE_NOTES.md`](screens/SCREEN_CAPTURE_NOTES.md).

A `main` pública contém apenas os assets que foram publicados em PRs documentais já aprovadas. Capturas que existem somente na branch de evolução agentic devem ser rotuladas como desenvolvimento e não devem ser apresentadas como funcionalidade integrada à `main`.

## Mockups conceituais

As imagens em [`mockups/`](mockups/) são conceitos visuais:

- `configuration-mockup.png` comunica uma direção futura para configuração, conectores, segurança e publicação;
- `builder-mockup.png` comunica a direção do editor visual e do fluxo de publicação;
- `mobile-mockup.png` comunica a direção do companion mobile.

Elas não substituem uma implementação, teste E2E, build assinado ou validação de dispositivo. Consulte [`mockups/MOCKUP_NOTES.md`](mockups/MOCKUP_NOTES.md).

## Assets herdados de documentação de integração

Os arquivos que permanecem diretamente sob `docs/images/`, inclusive imagens e ícones com nomes de Cline, Codex, Goose, IntelliJ, Marimo, n8n, Onyx, VS Code, Xcode, Zed e outros produtos, são **assets de documentação upstream ou de terceiros**. Eles ilustram como o ecossistema Ollama pode ser conectado a ferramentas externas; não são capturas da UI Classe A+ e não devem ser usados como evidência de que essas ferramentas estão embutidas, autenticadas ou homologadas neste fork.

Esses arquivos não foram removidos automaticamente porque páginas em `docs/integrations/`, capacidades e configurações do site podem referenciá-los. Antes de remover ou mover qualquer um, a alteração deve:

1. localizar referências no conteúdo, rotas e configuração do site;
2. preservar avisos de licença e atribuição aplicáveis;
3. atualizar todos os caminhos em uma mudança coesa;
4. executar validação de Markdown/JSON/YAML, integridade Classe A+ e build documental;
5. registrar o impacto no PR e confirmar que nenhuma página upstream ficou com imagem quebrada.

Arquivos sem referência foram apenas classificados como candidatos de limpeza. A ausência de uma referência textual não é autorização para apagar um asset sem revisar a origem e a atribuição.

## Política de marca e upstream

O projeto é um fork público baseado em Ollama e preserva a licença MIT, avisos e atribuições exigidos. A identidade Classe A+ deve ser comunicada por texto, documentação própria e capturas próprias, sem copiar UI, código ou assets proprietários de Manus, Claude, Codex, OmniRoute ou qualquer outro produto. A política operacional completa está em [`../../UPSTREAM_POLICY.md`](../../UPSTREAM_POLICY.md).

Quando uma página precisar explicar uma integração upstream, ela deve dizer explicitamente que se trata de documentação de compatibilidade ou de um serviço externo. Quando uma página apresentar uma tela do produto, deve apontar para `screens/` e para as notas de captura correspondentes.
