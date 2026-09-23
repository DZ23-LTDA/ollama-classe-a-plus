# Evidências oficiais — Composio, xAI/Grok, TikTok Shop e comércio social

Consulta: 2026-09-22.

## Composio

Fontes: [Composio docs](https://docs.composio.dev/docs), [toolkits](https://composio.dev/toolkits), [Composio Connect](https://docs.composio.dev/docs/composio-connect).

A documentação descreve SDK e MCP. O modo recomendado para aplicações é uma sessão por usuário (`composio.create(user_id)`), que fornece descoberta de tools, autenticação gerenciada e versionamento de toolkits. A sessão MCP expõe `session.mcp.url` e `session.mcp.headers`. O Composio Connect público usa `https://connect.composio.dev/mcp`, com autorização OAuth sob demanda e sete meta-tools para buscar tools, obter schemas, executar em paralelo, gerenciar conexões e aguardar autorização. A página de toolkits consultada informa 1.549 toolkits e categorias de Marketing/Social Media, E-commerce, Advertising/Marketing, CRM e Sales/Customer Support; o catálogo inclui Instagram, YouTube, WhatsApp, Twitter, Shopify e outros, mas presença no catálogo não prova que cada ação está habilitada, aprovada ou disponível para uma conta/região.

## xAI / Grok

Fontes: [xAI API overview](https://docs.x.ai/overview), [Grok Bot overview](https://docs.x.ai/grok-bot/overview), [Grok Bot](https://x.ai/bot).

A API oficial xAI expõe `https://api.x.ai/v1/responses` com bearer `XAI_API_KEY`, Responses API, tool calling, web search, structured outputs e modelos para texto/código; a página também descreve Voice API (STT/TTS/real-time), Imagine API para imagem/vídeo e Grok Build para coding agent. O Grok Bot é um produto hospedado separado: bots persistentes em computador cloud com browser, filesystem e terminal, colaboração entre bots, skills/routines e acesso desktop/mobile. Portanto o Classe A+ pode integrar a API xAI por provider OpenAI-compatible/Responses e implementar jornadas equivalentes com seus próprios runtime/browser/approvals, mas não deve declarar que possui o Grok Bot ou seus computadores cloud internos.

## TikTok Shop

Fontes: [TTS API overview](https://partner.tiktokshop.com/docv2/page/tts-api-concepts-overview), [Affiliate Seller API overview](https://partner.tiktokshop.com/docv2/page/affiliate-seller-api-overview), [TikTok developer blog](https://developers.tiktok.com/blog/2024-tiktok-shop-affiliate-apis-launch-developer-opportunity).

A TTS API oficial oferece autorização, webhooks e sandbox/testing tool, com Product, Order, Fulfillment, Return/Refund/Cancel, Logistics, Promotion, Finance, Seller, Authorization e Events APIs. A Affiliate API possui modos Seller, Creator e Partner para colaborações abertas/alvo, campanhas e tracking de conversões. A documentação informa que Affiliate APIs não estão disponíveis no Reino Unido e União Europeia e que não existe onboarding/moderação de creators totalmente programático para parceiros. A autorização varia entre seller, creator e partner e os escopos devem ser aprovados por categoria/mercado.

## Instagram e Shopify

Fontes: [Instagram Content Publishing](https://developers.facebook.com/documentation/instagram-platform/content-publishing), [Shopify Admin GraphQL API](https://shopify.dev/docs/api/admin-graphql/latest).

A API oficial do Instagram permite publicar imagens, vídeos, Reels e carrosséis em contas profissionais, com Meta Login, tokens, permissões específicas, mídia em URL pública e limite de publicação documentado. A página consultada diz que shopping tags não são suportadas nesse fluxo. O Shopify Admin GraphQL API permite construir apps que estendem o Admin, administrar produtos e pedidos mediante OAuth/access scopes e usa endpoint versionado; a API é rate-limited e erros de acesso podem aparecer no corpo GraphQL mesmo com HTTP 200.

## Estado do Classe A+

O repositório atual possui Growth OS sandbox e conectores HTTP/MCP genéricos, mas não possui ainda um conector nomeado e validado para Composio, xAI, TikTok Shop, Instagram ou Shopify. A implementação correta deve manter tokens server-side, OAuth por tenant, scopes mínimos, webhooks verificados, idempotência, approval para publicação/gasto/pedido/fulfillment, DLP, auditoria, rate limits, sandbox do provedor e testes por região. Nenhuma conta ou credencial externa foi declarada como conectada nesta consulta.
