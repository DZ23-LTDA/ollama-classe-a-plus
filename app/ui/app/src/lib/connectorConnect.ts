import { API_BASE } from "@/lib/config";

async function failure(response: Response, fallback: string): Promise<Error> {
  const text = (await response.text().catch(() => "")).trim();
  try {
    const parsed = JSON.parse(text) as { error?: string };
    if (parsed.error) return new Error(parsed.error);
  } catch {
    // Not JSON; use the raw text.
  }
  return new Error(text || fallback);
}

export async function connectWithKey(id: string, key: string): Promise<void> {
  const trimmed = key.trim();
  if (!trimmed) throw new Error("Cole a chave de API");
  const response = await fetch(
    `${API_BASE}/api/v1/connectors/${encodeURIComponent(id)}/key`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ key: trimmed }),
    },
  );
  if (!response.ok) throw await failure(response, "Falha ao conectar");
}

export async function disconnectConnector(id: string): Promise<void> {
  const response = await fetch(
    `${API_BASE}/api/v1/connectors/${encodeURIComponent(id)}/key`,
    {
      method: "DELETE",
    },
  );
  if (!response.ok) throw await failure(response, "Falha ao desconectar");
}

// Where each service issues API keys, so the user does not have to search.
const KEY_HELP: Record<string, string> = {
  airtable: "https://airtable.com/create/tokens",
  betterstack: "https://uptime.betterstack.com/team/api-tokens",
  clerk: "https://dashboard.clerk.com/last-active?path=api-keys",
  cloudflare: "https://dash.cloudflare.com/profile/api-tokens",
  digitalocean: "https://cloud.digitalocean.com/account/api/tokens",
  github: "https://github.com/settings/personal-access-tokens",
  hubspot: "https://app.hubspot.com/private-apps",
  "mercado-pago": "https://www.mercadopago.com.br/developers/panel/app",
  neon: "https://console.neon.tech/app/settings/api-keys",
  netlify: "https://app.netlify.com/user/applications#personal-access-tokens",
  posthog: "https://us.posthog.com/settings/user-api-keys",
  render: "https://dashboard.render.com/u/settings#api-keys",
  resend: "https://resend.com/api-keys",
  sendgrid: "https://app.sendgrid.com/settings/api_keys",
  sentry: "https://sentry.io/settings/account/api/auth-tokens/",
  stripe: "https://dashboard.stripe.com/apikeys",
  supabase: "https://supabase.com/dashboard/account/tokens",
  vercel: "https://vercel.com/account/settings/tokens",
};

export function keyHelpURL(id: string): string | undefined {
  return KEY_HELP[id];
}
