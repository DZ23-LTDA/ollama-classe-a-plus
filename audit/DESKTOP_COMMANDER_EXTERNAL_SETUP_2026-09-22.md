# Evidência externa — Remote Desktop Commander

Fonte consultada: [Remote Desktop Commander setup](https://github.com/desktop-commander/remote-desktop-commander/blob/main/docs/SETUP.md), consultada em 2026-09-22.

A documentação oficial descreve duas etapas. Primeiro, no computador que será controlado, o operador executa `npx @wonderwhy-er/desktop-commander@latest remote`; o comando abre uma página de verificação de dispositivo e imprime o mesmo código no terminal. O operador entra com Google ou e-mail, compara os códigos e confirma `Verify Device`. O terminal precisa permanecer aberto enquanto o dispositivo estiver conectado.

Depois, o assistente configura o servidor `https://mcp.desktopcommander.app/mcp`. O fluxo abre uma autorização OAuth e deve usar a mesma conta do pareamento. A página também informa que o dashboard lista dispositivos, status online, último contato e ações de adicionar ou revogar dispositivos.

A evidência confirma que o teste real exige uma conta do Desktop Commander, um computador com Node.js 18 ou superior e autorização do operador. O sandbox não possui credenciais, pareamento ou dispositivo externo; portanto esta rodada pode implementar e testar o adapter local, mas não pode declarar uma conexão remota autenticada.
