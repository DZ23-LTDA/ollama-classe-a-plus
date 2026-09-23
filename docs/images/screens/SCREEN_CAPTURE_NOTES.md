# Notas das capturas

## Escopo do diretório

Estas notas cobrem somente as capturas Classe A+ em `docs/images/screens/`. A raiz `docs/images/` também contém imagens herdadas da documentação de integrações Ollama e de terceiros; elas não são telas deste produto. A classificação completa está em [`../ASSET_PROVENANCE.md`](../ASSET_PROVENANCE.md).

## Captura publicada na main

`agentic-console.png` é uma captura real da rota `/agentic` da UI React atual, renderizada com Chromium e dados demonstrativos interceptados localmente. Ela mostra o Mission Console, criação de missão, métricas, orquestração multiagente, pesquisa profunda, timeline e approval pendente.

Esta captura foi publicada na `main` como documentação do estado visual observado durante o desenvolvimento. Ela não representa um release instalável e não comprova credenciais, contas OAuth, providers externos, dispositivos físicos, deploy, distribuição em lojas ou homologação.

## Capturas da branch de evolução

A série `class-a-plus-*.png` e o alias `settings.png` representam a evolução visual mais recente do shell Classe A+. Quando esses arquivos forem adicionados a uma branch ou PR, a legenda deve dizer explicitamente se a captura pertence à branch de desenvolvimento e se o código correspondente ainda não está integrado à `main`. Não se deve usar uma imagem da branch de evolução para afirmar que uma funcionalidade já foi entregue na linha pública.

## Método e limites

As capturas devem ser feitas com `app/ui/app/scripts/capture-parity-screens.mjs`, usando Chromium, o Vite dev e o servidor Ollama local em loopback. O capturador deve falhar fechado diante de uma página quase vazia, erro de console, asset ausente ou resposta inesperada da API.

Screenshots reais demonstram somente a implementação e o estado observado no ambiente em que foram capturados. Eles não substituem testes de integração, smoke com contas externas, staging distribuído, validação em dispositivos ou evidência de release. Mockups conceituais continuam documentados separadamente em `docs/images/mockups/MOCKUP_NOTES.md`.
