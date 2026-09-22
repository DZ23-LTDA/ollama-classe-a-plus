import { Link } from "@tanstack/react-router";
import {
  ArrowRightIcon,
  CheckCircleIcon,
  ClockIcon,
  FolderOpenIcon,
  LockClosedIcon,
  PlusIcon,
  SparklesIcon,
} from "@heroicons/react/24/outline";
import { AppSidebar, type AppSection } from "@/components/AppSidebar";
import { SidebarLayout } from "@/components/layout/layout";

type ProductPageKind = "library" | "projects" | "scheduled" | "skills" | "plugins" | "tasks";

const pageCopy: Record<
  ProductPageKind,
  { title: string; eyebrow: string; description: string; action: string; metric: string; empty: string }
> = {
  library: {
    title: "Biblioteca",
    eyebrow: "Artifacts e arquivos",
    description: "Encontre documentos, sites, jogos, dashboards e outros artifacts produzidos pelas suas missões.",
    action: "Criar artifact",
    metric: "0 artifacts nesta sessão",
    empty: "Os artifacts gerados aparecerão aqui com preview, hash, versão e download.",
  },
  projects: {
    title: "Projetos",
    eyebrow: "Contexto persistente",
    description: "Organize memória, fontes, tarefas, membros e builders em workspaces isolados.",
    action: "Novo projeto",
    metric: "0 projetos ativos",
    empty: "Crie um projeto para manter instruções, arquivos e contexto entre missões.",
  },
  scheduled: {
    title: "Agendado",
    eyebrow: "Automação controlada",
    description: "Acompanhe tarefas recorrentes, webhooks, retries e execuções aguardando aprovação.",
    action: "Agendar tarefa",
    metric: "0 execuções pendentes",
    empty: "Nenhuma automação está agendada. As tarefas futuras serão listadas com estado e próxima execução.",
  },
  skills: {
    title: "Habilidades",
    eyebrow: "Skills e MCP",
    description: "Instale capacidades com manifesto, versão, escopos e aprovação por ferramenta.",
    action: "Adicionar habilidade",
    metric: "Runtime protegido",
    empty: "Nenhuma habilidade adicional foi habilitada neste workspace.",
  },
  plugins: {
    title: "Plugins",
    eyebrow: "Conectores e providers",
    description: "Conecte GitHub, Google Workspace, Claude, Codex, OmniRoute e outros serviços sem expor segredos.",
    action: "Adicionar conexão",
    metric: "Credenciais ocultas",
    empty: "Nenhuma conexão foi configurada. Tokens serão mantidos no ambiente seguro do operador.",
  },
  tasks: {
    title: "Tarefas",
    eyebrow: "Inbox de missões",
    description: "Veja missões em execução, aguardando approval, concluídas ou em recuperação.",
    action: "Nova tarefa",
    metric: "0 missões em execução",
    empty: "As missões criadas no Agentic Console aparecerão aqui com timeline e artifacts.",
  },
};

function StatCard({ label, value, tone = "neutral" }: { label: string; value: string; tone?: "neutral" | "green" | "violet" }) {
  const toneClass =
    tone === "green"
      ? "bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-300"
      : tone === "violet"
        ? "bg-violet-50 text-violet-700 dark:bg-violet-950/30 dark:text-violet-300"
        : "bg-neutral-100 text-neutral-700 dark:bg-neutral-800 dark:text-neutral-300";
  return (
    <div className="rounded-2xl border border-neutral-200/80 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900">
      <div className="text-[11px] uppercase tracking-[0.12em] text-neutral-400">{label}</div>
      <div className={`mt-3 inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${toneClass}`}>{value}</div>
    </div>
  );
}

export function ProductWorkspacePage({ kind }: { kind: ProductPageKind }) {
  const copy = pageCopy[kind];
  const current = kind as AppSection;

  return (
    <SidebarLayout title={copy.title} sidebar={<AppSidebar current={current} />}>
      <div className="min-h-0 flex-1 overflow-y-auto bg-neutral-50 dark:bg-neutral-900">
        <div className="mx-auto w-full max-w-6xl px-6 pb-14 pt-10 lg:px-12">
          <div className="flex flex-col gap-5 border-b border-neutral-200 pb-8 dark:border-neutral-800 md:flex-row md:items-end md:justify-between">
            <div className="max-w-2xl">
              <div className="mb-3 flex items-center gap-2 text-xs font-medium text-violet-600 dark:text-violet-300">
                <SparklesIcon className="h-4 w-4" />
                {copy.eyebrow}
              </div>
              <h2 className="font-rounded text-3xl font-semibold tracking-tight text-neutral-950 dark:text-white">{copy.title}</h2>
              <p className="mt-3 text-sm leading-6 text-neutral-500 dark:text-neutral-400">{copy.description}</p>
            </div>
            <button className="inline-flex items-center justify-center gap-2 rounded-xl bg-neutral-950 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition hover:bg-neutral-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-400 dark:bg-white dark:text-neutral-950 dark:hover:bg-neutral-200">
              <PlusIcon className="h-4 w-4" />
              {copy.action}
            </button>
          </div>

          <div className="mt-7 grid gap-3 sm:grid-cols-3">
            <StatCard label="Status" value={copy.metric} tone="green" />
            <StatCard label="Segurança" value="Approvals ativos" tone="violet" />
            <StatCard label="Persistência" value="Local-first" />
          </div>

          <div className="mt-7 grid gap-5 lg:grid-cols-[1.45fr_0.8fr]">
            <section className="rounded-2xl border border-neutral-200/80 bg-white p-6 dark:border-neutral-800 dark:bg-neutral-900">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-semibold text-neutral-900 dark:text-white">Estado do workspace</h3>
                  <p className="mt-1 text-xs text-neutral-500 dark:text-neutral-400">Dados reais aparecerão aqui quando o fluxo vertical estiver conectado.</p>
                </div>
                <CheckCircleIcon className="h-5 w-5 text-emerald-500" />
              </div>
              <div className="mt-8 rounded-xl border border-dashed border-neutral-300 px-6 py-10 text-center dark:border-neutral-700">
                <FolderOpenIcon className="mx-auto h-8 w-8 text-neutral-300 dark:text-neutral-600" />
                <p className="mx-auto mt-3 max-w-md text-sm text-neutral-500 dark:text-neutral-400">{copy.empty}</p>
                <button className="mt-5 inline-flex items-center gap-2 rounded-lg border border-neutral-200 px-3 py-2 text-xs font-medium text-neutral-700 hover:bg-neutral-50 dark:border-neutral-700 dark:text-neutral-200 dark:hover:bg-neutral-800">
                  Explorar fluxo
                  <ArrowRightIcon className="h-3.5 w-3.5" />
                </button>
              </div>
            </section>

            <aside className="space-y-5">
              <section className="rounded-2xl border border-neutral-200/80 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
                <div className="flex items-center gap-2 text-sm font-semibold text-neutral-900 dark:text-white">
                  <LockClosedIcon className="h-4 w-4 text-emerald-500" />
                  Política ativa
                </div>
                <ul className="mt-4 space-y-3 text-xs leading-5 text-neutral-500 dark:text-neutral-400">
                  <li className="flex gap-2"><CheckCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />Secrets não aparecem na interface.</li>
                  <li className="flex gap-2"><CheckCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />Ações externas exigem approval.</li>
                  <li className="flex gap-2"><CheckCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />Local-only não usa fallback remoto.</li>
                </ul>
              </section>
              <section className="rounded-2xl border border-neutral-200/80 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
                <div className="flex items-center gap-2 text-sm font-semibold text-neutral-900 dark:text-white">
                  <ClockIcon className="h-4 w-4 text-violet-500" />
                  Próximos passos
                </div>
                <p className="mt-3 text-xs leading-5 text-neutral-500 dark:text-neutral-400">Conecte esta superfície ao contrato da API agentic para substituir o estado vazio por dados persistidos.</p>
                <Link to="/agentic" className="mt-4 inline-flex items-center gap-2 text-xs font-medium text-violet-600 hover:text-violet-500 dark:text-violet-300">
                  Abrir Agentic Console <ArrowRightIcon className="h-3.5 w-3.5" />
                </Link>
              </section>
            </aside>
          </div>
        </div>
      </div>
    </SidebarLayout>
  );
}
