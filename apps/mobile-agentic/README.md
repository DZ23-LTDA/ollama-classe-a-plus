# DZ23 Agentic Mobile

Cliente Expo para acompanhar missões, visualizar timeline, aprovar/rejeitar passos protegidos e iniciar execução pelo mesmo endpoint agentic v1 do Ollama DZ23.

## Execução

```bash
npm install
npx expo start
```

No emulador ou dispositivo físico, informe em **Servidor** uma URL alcançável pelo celular, por exemplo `http://192.168.0.10:11434`. O cliente persiste apenas essa URL localmente; tokens e segredos permanecem no servidor e não são enviados para o bundle mobile.

O servidor continua sendo a autoridade de policy. O mobile não executa shell, browser, processos ou conectores diretamente; ele apenas solicita ações pela API, observa eventos e apresenta approvals.
