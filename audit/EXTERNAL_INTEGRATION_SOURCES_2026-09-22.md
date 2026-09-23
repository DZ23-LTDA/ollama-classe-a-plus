# Fontes externas de integração — 2026-09-22

## Woovi / OpenPix

A documentação oficial da Woovi descreve uma API REST Pix, ambiente de teste, chaves de API e webhooks em tempo real. A página inicial mostra `POST https://api.woovi.com/api/v1/charge` com o header `Authorization` usando o App ID e um payload com `correlationID` e `value`.

A documentação oficial de eventos lista, entre outros, `OPENPIX:CHARGE_CREATED`, `OPENPIX:CHARGE_COMPLETED`, `OPENPIX:CHARGE_EXPIRED`, `OPENPIX:TRANSACTION_RECEIVED`, eventos de estorno, movimentação, disputa e Pix Automático. O adapter do projeto deve validar assinatura/headers conforme a documentação da conta, deduplicar por evento/correlation ID, limitar payload, registrar auditoria e manter criação de cobranças, estornos e pagamentos sujeitos a approval.

Fontes oficiais consultadas:

- [Woovi Developers](https://developers.woovi.com/en/)
- [Woovi Webhook Event Types](https://developers.woovi.com/en/docs/webhook/webhook-events-type)

## Fiscal

O connector nativo chamado `Fiscal` disponível na sessão aponta para `https://api.fiscal.ai/mcp` e descreve dados financeiros, filings, preços, ownership, notícias e análises de empresas. Ele **não é um emissor de NF-e/NFS-e**. A emissão fiscal permanece um adapter separado, dependente da escolha do provedor brasileiro, credenciais da empresa, certificado quando exigido, ambiente de homologação, município/UF e regras fiscais do operador.

## Limites

A existência de uma entrada no catálogo não significa conta autorizada, credencial válida, homologação fiscal, capacidade de emissão ou permissão para movimentar dinheiro. Nenhuma cobrança, pagamento, estorno, anúncio, pedido, deploy ou nota fiscal deve ser executado sem credencial do operador, approval aplicável e evidência de resposta real.
