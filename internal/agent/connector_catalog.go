package agent

import "sort"

type ConnectorCatalogEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Kind        string   `json:"kind"`
	Description string   `json:"description"`
	Auth        string   `json:"auth"`
	Source      string   `json:"source"`
	Status      string   `json:"status"`
	Scopes      []string `json:"scopes,omitempty"`
}

// ConnectorCatalog is intentionally a capability catalog, not a claim that every
// provider is connected. Entries requiring an operator credential remain explicit.
func ConnectorCatalog() []ConnectorCatalogEntry {
	entries := []ConnectorCatalogEntry{
		{ID: "github", Name: "GitHub", Category: "Desenvolvimento", Kind: "app", Description: "Repositórios, branches, issues, pull requests, Actions e releases.", Auth: "oauth", Source: "built-in/example", Status: "available", Scopes: []string{"repositories", "issues", "pull_requests", "actions"}},
		{ID: "google-workspace", Name: "Google Workspace", Category: "Produtividade", Kind: "app", Description: "Gmail, Drive, Docs, Sheets, Calendar e contatos.", Auth: "oauth", Source: "built-in", Status: "available", Scopes: []string{"gmail", "drive", "docs", "sheets", "calendar"}},
		{ID: "composio", Name: "Composio", Category: "Automação", Kind: "mcp", Description: "Gateway MCP para aplicativos autorizados pelo operador, com descoberta de ferramentas.", Auth: "oauth_or_api_key", Source: "mcp", Status: "available", Scopes: []string{"tool_discovery", "oauth", "external_actions"}},
		{ID: "slack", Name: "Slack", Category: "Comunicação", Kind: "app", Description: "Mensagens, canais, threads e notificações com aprovação para efeitos externos.", Auth: "oauth", Source: "built-in/example", Status: "available", Scopes: []string{"messages", "channels"}},
		{ID: "discord", Name: "Discord", Category: "Comunicação", Kind: "custom_api", Description: "Bots, canais e mensagens via aplicação Discord ou Composio.", Auth: "bot_or_oauth", Source: "custom_api/mcp", Status: "operator_setup_required", Scopes: []string{"messages", "channels"}},
		{ID: "whatsapp", Name: "WhatsApp Business", Category: "Comunicação", Kind: "custom_api", Description: "Mensagens transacionais pela WhatsApp Cloud API/Meta Business.", Auth: "api_key", Source: "custom_api", Status: "operator_setup_required", Scopes: []string{"messages", "templates"}},
		{ID: "notion", Name: "Notion", Category: "Produtividade", Kind: "app", Description: "Páginas, databases, blocos e conhecimento operacional.", Auth: "oauth_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"pages", "databases"}},
		{ID: "jira", Name: "Jira", Category: "Desenvolvimento", Kind: "custom_api", Description: "Projetos, issues, sprints e workflow de engenharia.", Auth: "oauth_or_api_key", Source: "custom_api/mcp", Status: "operator_setup_required", Scopes: []string{"issues", "projects"}},
		{ID: "hubspot", Name: "HubSpot", Category: "CRM e Vendas", Kind: "app", Description: "CRM, contatos, empresas, negócios e automações comerciais.", Auth: "oauth", Source: "built-in", Status: "available", Scopes: []string{"crm", "contacts"}},
		{ID: "stripe", Name: "Stripe", Category: "Pagamentos", Kind: "app", Description: "Clientes, produtos, invoices e webhooks; movimentações sensíveis permanecem approval-gated.", Auth: "oauth_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"customers", "products", "invoices", "webhooks"}},
		{ID: "woovi-openpix", Name: "Woovi / OpenPix", Category: "Pagamentos", Kind: "custom_api", Description: "Cobranças Pix, QR Code, conciliação e webhooks de pagamento em tempo real.", Auth: "api_key", Source: "custom_api", Status: "operator_setup_required", Scopes: []string{"charges", "webhooks", "transactions"}},
		{ID: "fiscal-invoicing", Name: "Emissão fiscal NF-e/NFS-e", Category: "Fiscal", Kind: "custom_api", Description: "Contrato para emissão, consulta, cancelamento e armazenamento de XML/DANFE por provedor fiscal homologado.", Auth: "api_key_or_certificate", Source: "custom_api", Status: "provider_selection_required", Scopes: []string{"nfe", "nfse", "xml", "danfe", "webhooks"}},
		{ID: "fiscal-ai", Name: "Fiscal AI", Category: "Dados financeiros", Kind: "mcp", Description: "Dados financeiros e análises de empresas; não é emissor de NF-e/NFS-e.", Auth: "oauth_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"financial_data", "filings"}},
		{ID: "instagram", Name: "Instagram", Category: "Marketing", Kind: "app", Description: "Contas profissionais, conteúdo e métricas via permissões da Meta.", Auth: "oauth", Source: "built-in", Status: "available", Scopes: []string{"content", "insights"}},
		{ID: "meta-ads", Name: "Meta Ads Manager", Category: "Marketing", Kind: "app", Description: "Campanhas, anúncios, orçamento e métricas com approval para publicação e gasto.", Auth: "oauth", Source: "built-in", Status: "available", Scopes: []string{"campaigns", "insights", "ads"}},
		{ID: "tiktok-business", Name: "TikTok for Business", Category: "Marketing", Kind: "app", Description: "Campanhas e métricas TikTok; TikTok Shop depende de elegibilidade e APIs da conta.", Auth: "oauth", Source: "built-in", Status: "available", Scopes: []string{"campaigns", "insights"}},
		{ID: "shopify", Name: "Shopify", Category: "Comércio", Kind: "app", Description: "Produtos, inventário, pedidos e fulfillment com aprovação antes de efeitos comerciais.", Auth: "oauth_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"products", "inventory", "orders"}},
		{ID: "woocommerce", Name: "WooCommerce", Category: "Comércio", Kind: "custom_api", Description: "Catálogo, pedidos e estoque via REST API do operador.", Auth: "api_key", Source: "custom_api", Status: "operator_setup_required", Scopes: []string{"products", "orders", "inventory"}},
		{ID: "mercado-livre", Name: "Mercado Livre", Category: "Comércio", Kind: "custom_api", Description: "Catálogo, pedidos, anúncios e logística conforme permissões da conta.", Auth: "oauth", Source: "custom_api", Status: "operator_setup_required", Scopes: []string{"catalog", "orders", "shipping"}},
		{ID: "amazon-seller", Name: "Amazon Seller", Category: "Comércio", Kind: "custom_api", Description: "Produtos, pedidos e fulfillment pela Selling Partner API.", Auth: "oauth", Source: "custom_api", Status: "operator_setup_required", Scopes: []string{"catalog", "orders", "inventory"}},
		{ID: "vercel", Name: "Vercel", Category: "Deploy", Kind: "app", Description: "Projetos, deployments, domínios e logs de publicação.", Auth: "oauth_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"projects", "deployments", "domains"}},
		{ID: "netlify", Name: "Netlify", Category: "Deploy", Kind: "app", Description: "Sites, deploys, domains e logs de publicação.", Auth: "oauth_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"sites", "deploys", "domains"}},
		{ID: "cloudflare", Name: "Cloudflare", Category: "Deploy", Kind: "app", Description: "Workers, Pages, DNS e configurações de edge com escopo explícito.", Auth: "oauth_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"pages", "workers", "dns"}},
		{ID: "zapier", Name: "Zapier", Category: "Automação", Kind: "mcp", Description: "Ações e workflows externos através de conexão autorizada.", Auth: "oauth_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"actions", "triggers"}},
		{ID: "n8n", Name: "n8n", Category: "Automação", Kind: "mcp", Description: "Workflows self-hosted ou cloud para integrações operacionais.", Auth: "url_or_api_key", Source: "built-in", Status: "available", Scopes: []string{"workflows", "executions"}},
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}
