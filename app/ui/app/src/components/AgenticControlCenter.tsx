import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowTopRightOnSquareIcon,
  CheckCircleIcon,
  CircleStackIcon,
  ExclamationTriangleIcon,
  KeyIcon,
  LockClosedIcon,
  WrenchScrewdriverIcon,
} from "@heroicons/react/24/outline";
import { API_BASE } from "@/lib/config";
import { useModels } from "@/hooks/useModels";

type SafeConfig = {
  runtime?: string;
  store?: string;
  queue?: string;
  auth_required?: boolean;
  approval_gated_tools?: boolean;
  workspace_isolation?: boolean;
  planner_model_configured?: boolean;
  embedding_configured?: boolean;
  connectors_configured?: boolean;
  mcp_configured?: boolean;
  media_configured?: boolean;
  deployments_configured?: boolean;
  otlp_configured?: boolean;
  push_configured?: boolean;
};

function StatusPill({ ok, children }: { ok: boolean; children: React.ReactNode }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2 py-1 text-[10px] font-medium ${
        ok
          ? "bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-300"
          : "bg-neutral-100 text-neutral-500 dark:bg-neutral-800 dark:text-neutral-400"
      }`}
    >
      {ok ? <CheckCircleIcon className="h-3.5 w-3.5" /> : <ExclamationTriangleIcon className="h-3.5 w-3.5" />}
      {children}
    </span>
  );
}

export function AgenticControlCenter() {
  const { data: models = [], isLoading: modelsLoading } = useModels();
  const { data: safeConfig, isLoading: configLoading } = useQuery<SafeConfig>({
    queryKey: ["agent-safe-config"],
    queryFn: async () => {
      const response = await fetch(`${API_BASE}/api/agent/v1/config/safe`);
      if (!response.ok) throw new Error(`agent config ${response.status}`);
      return (await response.json()) as SafeConfig;
    },
    retry: false,
  });

  const providers = useMemo(() => {
    const values = new Set<string>();
    for (const model of models) {
      values.add(model.provider || (model.kind === "router" ? "DZ23 Router" : "Ollama"));
    }
    return [...values].sort();
  }, [models]);

  const runtimeReady = Boolean(safeConfig?.runtime);
  const statusRows = [
    { label: "Planner", enabled: safeConfig?.planner_model_configured ?? false },
    { label: "Connectors", enabled: safeConfig?.connectors_configured ?? false },
    { label: "MCP / skills", enabled: safeConfig?.mcp_configured ?? false },
    { label: "Media provider", enabled: safeConfig?.media_configured ?? false },
    { label: "Deploy providers", enabled: safeConfig?.deployments_configured ?? false },
    { label: "OTLP traces", enabled: safeConfig?.otlp_configured ?? false },
  ];

  return (
    <section id="agentic" className="overflow-hidden rounded-xl border border-neutral-200/80 bg-white dark:border-neutral-800 dark:bg-neutral-800">
      <div className="border-b border-neutral-200/80 bg-gradient-to-br from-neutral-50 to-violet-50/40 p-5 dark:border-neutral-700 dark:from-neutral-900 dark:to-violet-950/20">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <div className="flex items-center gap-2 text-sm font-semibold text-neutral-900 dark:text-white">
              <WrenchScrewdriverIcon className="h-5 w-5 text-violet-500" />
              Agentic Control Center
            </div>
            <p className="mt-1 max-w-2xl text-xs leading-5 text-neutral-500 dark:text-neutral-400">
              Estado sanitizado do runtime, providers, approvals e integrações. Tokens, caminhos privados e ciphertext nunca são retornados.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <StatusPill ok={runtimeReady}>{configLoading ? "Consultando" : runtimeReady ? "Runtime online" : "Local mode"}</StatusPill>
            <a href="/agentic" className="inline-flex items-center gap-1 text-xs font-medium text-violet-600 hover:text-violet-500 dark:text-violet-300">
              Abrir console <ArrowTopRightOnSquareIcon className="h-3.5 w-3.5" />
            </a>
          </div>
        </div>
      </div>

      <div className="grid gap-4 p-5 lg:grid-cols-[1.1fr_0.9fr]">
        <div>
          <div className="mb-3 flex items-center gap-2 text-xs font-semibold text-neutral-800 dark:text-neutral-200">
            <CircleStackIcon className="h-4 w-4 text-neutral-400" />
            Providers visíveis no catálogo
          </div>
          <div className="flex flex-wrap gap-2">
            {modelsLoading ? (
              <span className="text-xs text-neutral-400">Carregando catálogo…</span>
            ) : providers.length > 0 ? (
              providers.map((provider) => (
                <span key={provider} className="rounded-lg border border-neutral-200 px-2.5 py-1.5 text-xs text-neutral-600 dark:border-neutral-700 dark:text-neutral-300">
                  {provider}
                </span>
              ))
            ) : (
              <span className="text-xs text-neutral-400">Nenhum provider remoto configurado; modelos locais continuam disponíveis.</span>
            )}
          </div>
          <div className="mt-4 flex flex-wrap gap-2">
            <StatusPill ok={Boolean(safeConfig?.approval_gated_tools)}>Approvals server-side</StatusPill>
            <StatusPill ok={Boolean(safeConfig?.workspace_isolation)}>Workspace isolado</StatusPill>
            <StatusPill ok={safeConfig?.auth_required ?? false}>{safeConfig?.auth_required ? "Auth obrigatória" : "Auth local"}</StatusPill>
          </div>
        </div>

        <div className="rounded-xl border border-neutral-200/80 p-4 dark:border-neutral-700">
          <div className="flex items-center gap-2 text-xs font-semibold text-neutral-800 dark:text-neutral-200">
            <KeyIcon className="h-4 w-4 text-neutral-400" />
            Capacidades configuradas
          </div>
          <div className="mt-3 grid grid-cols-2 gap-2">
            {statusRows.map((row) => (
              <div key={row.label} className="flex items-center gap-2 text-[11px] text-neutral-500 dark:text-neutral-400">
                <span className={`h-1.5 w-1.5 rounded-full ${row.enabled ? "bg-emerald-500" : "bg-neutral-300 dark:bg-neutral-600"}`} />
                {row.label}
              </div>
            ))}
          </div>
          <div className="mt-4 flex items-start gap-2 rounded-lg bg-neutral-50 p-3 text-[10px] leading-4 text-neutral-500 dark:bg-neutral-900 dark:text-neutral-400">
            <LockClosedIcon className="mt-0.5 h-3.5 w-3.5 shrink-0 text-emerald-500" />
            A edição de credenciais permanece fora da UI e deve usar ambiente seguro ou secrets manager.
          </div>
        </div>
      </div>
    </section>
  );
}
