import { Link } from "@tanstack/react-router";
import { useCallback, useEffect, useMemo, useState } from "react";
import { API_BASE } from "@/lib/config";
import {
  createProject,
  createSchedule,
  deleteProject,
  deleteSchedule,
  listCLIStatus,
  listConnectors,
  listMCPServers,
  listMissions,
  listProjects,
  listSchedules,
  listSkills,
  updateProject,
} from "@/lib/agenticClient";
import type { AgentConnector, AgentMCPServer, AgentMission, AgentProject, AgentSchedule, AgentSkill } from "@/lib/agenticClient";
import {
  ArrowRightIcon,
  CheckCircleIcon,
  ClockIcon,
  FolderOpenIcon,
  LockClosedIcon,
  PlusIcon,
  SparklesIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { AppSidebar, type AppSection } from "@/components/AppSidebar";
import { SidebarLayout } from "@/components/layout/layout";

type ProductPageKind = "library" | "projects" | "scheduled" | "skills" | "plugins" | "tasks";

const pageCopy: Record<ProductPageKind, { title: string; eyebrow: string; description: string; action: string }> = {
  library: { title: "Biblioteca", eyebrow: "Artifacts e arquivos", description: "Encontre documentos, sites, jogos, dashboards e outros artifacts produzidos pelas suas missões.", action: "Abrir Agentic Console" },
  projects: { title: "Projetos", eyebrow: "Contexto persistente", description: "Organize memória, fontes, tarefas, membros e builders em workspaces isolados.", action: "Novo projeto" },
  scheduled: { title: "Agendado", eyebrow: "Automação controlada", description: "Acompanhe tarefas recorrentes, webhooks, retries e execuções aguardando aprovação.", action: "Agendar tarefa" },
  skills: { title: "Habilidades", eyebrow: "Skills e MCP", description: "Veja manifestos, versões, escopos e confiança das habilidades carregadas pelo runtime.", action: "Abrir configuração" },
  plugins: { title: "Plugins", eyebrow: "Conectores e providers", description: "Inspecione GitHub, Google Workspace, Claude, Codex, OmniRoute, MCP e outros serviços sem expor segredos.", action: "Abrir configuração" },
  tasks: { title: "Tarefas", eyebrow: "Inbox de missões", description: "Veja missões em execução, aguardando approval, concluídas ou em recuperação.", action: "Nova tarefa" },
};

function StatCard({ label, value, tone = "neutral" }: { label: string; value: string; tone?: "neutral" | "green" | "violet" }) {
  const toneClass = tone === "green" ? "bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-300" : tone === "violet" ? "bg-violet-50 text-violet-700 dark:bg-violet-950/30 dark:text-violet-300" : "bg-neutral-100 text-neutral-700 dark:bg-neutral-800 dark:text-neutral-300";
  return <div className="rounded-2xl border border-neutral-200/80 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900"><div className="text-[11px] uppercase tracking-[0.12em] text-neutral-400">{label}</div><div className={`mt-3 inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${toneClass}`}>{value}</div></div>;
}

function EmptyState({ message, action, onAction }: { message: string; action?: string; onAction?: () => void }) {
  return <div className="rounded-xl border border-dashed border-neutral-300 px-6 py-10 text-center dark:border-neutral-700"><FolderOpenIcon className="mx-auto h-8 w-8 text-neutral-300 dark:text-neutral-600" /><p className="mx-auto mt-3 max-w-md text-sm text-neutral-500 dark:text-neutral-400">{message}</p>{action && onAction && <button onClick={onAction} className="mt-5 inline-flex items-center gap-2 rounded-lg border border-neutral-200 px-3 py-2 text-xs font-medium text-neutral-700 hover:bg-neutral-50 dark:border-neutral-700 dark:text-neutral-200 dark:hover:bg-neutral-800">{action}<ArrowRightIcon className="h-3.5 w-3.5" /></button>}</div>;
}

function ResourceRow({ children, onDelete }: { children: React.ReactNode; onDelete?: () => void }) {
  return <div className="flex items-center justify-between gap-3 rounded-xl border border-neutral-200/80 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900"><div className="min-w-0 flex-1">{children}</div>{onDelete && <button type="button" onClick={onDelete} className="rounded-lg p-2 text-neutral-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/30" aria-label="Excluir"><TrashIcon className="h-4 w-4" /></button>}</div>;
}

export function ProductWorkspacePage({ kind }: { kind: ProductPageKind }) {
  const copy = pageCopy[kind];
  const current = kind as AppSection;
  const [notice, setNotice] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [projects, setProjects] = useState<AgentProject[]>([]);
  const [missions, setMissions] = useState<AgentMission[]>([]);
  const [schedules, setSchedules] = useState<AgentSchedule[]>([]);
  const [skills, setSkills] = useState<AgentSkill[]>([]);
  const [connectors, setConnectors] = useState<AgentConnector[]>([]);
  const [mcpServers, setMcpServers] = useState<AgentMCPServer[]>([]);
  const [cliCount, setCliCount] = useState(0);
  const [projectName, setProjectName] = useState("");
  const [projectRoot, setProjectRoot] = useState("");
  const [scheduleObjective, setScheduleObjective] = useState("");
  const [scheduleInterval, setScheduleInterval] = useState("3600");
  const [scheduleProject, setScheduleProject] = useState("");
  const [editingProject, setEditingProject] = useState<string | null>(null);
  const [editingProjectName, setEditingProjectName] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      if (kind === "projects" || kind === "scheduled") setProjects((await listProjects()).projects);
      if (kind === "tasks" || kind === "library") setMissions((await listMissions()).missions);
      if (kind === "scheduled") setSchedules((await listSchedules()).schedules);
      if (kind === "skills") setSkills((await listSkills()).skills);
      if (kind === "plugins") {
        const [connectorResult, mcpResult, cliResult] = await Promise.all([listConnectors(), listMCPServers(), listCLIStatus()]);
        setConnectors(connectorResult.connectors);
        setMcpServers(mcpResult.servers);
        setCliCount(cliResult.tools.filter((item) => item.installed).length);
      }
    } catch (cause) {
      setNotice(cause instanceof Error ? cause.message : "Não foi possível carregar os dados persistidos.");
    } finally {
      setLoading(false);
    }
  }, [kind]);

  useEffect(() => { void refresh(); }, [refresh]);

  const artifacts = useMemo(() => missions.flatMap((mission) => (mission.artifacts ?? []).map((artifact) => ({ ...artifact, mission }))), [missions]);
  const runningMissions = missions.filter((mission) => ["RUNNING", "OBSERVING", "RECOVERING"].includes(mission.state)).length;

  const handleAction = () => {
    if (kind === "tasks" || kind === "library") window.location.assign("/agentic");
    else if (kind === "projects") document.getElementById("new-project-name")?.focus();
    else if (kind === "scheduled") document.getElementById("new-schedule-objective")?.focus();
    else window.location.assign("/settings#agentic");
  };

  const createNewProject = async () => {
    if (!projectName.trim()) return;
    try {
      await createProject(projectName, projectRoot);
      setProjectName(""); setProjectRoot(""); setNotice("Projeto criado e persistido no ContextStore."); await refresh();
    } catch (cause) { setNotice(cause instanceof Error ? cause.message : "Falha ao criar projeto."); }
  };

  const saveProject = async (project: AgentProject) => {
    if (!editingProjectName.trim()) return;
    try { await updateProject(project.id, editingProjectName, project.root ?? ""); setEditingProject(null); setNotice("Projeto atualizado."); await refresh(); }
    catch (cause) { setNotice(cause instanceof Error ? cause.message : "Falha ao atualizar projeto."); }
  };

  const removeProject = async (project: AgentProject) => {
    if (!window.confirm(`Excluir o projeto ${project.name}? Memórias locais associadas também serão removidas.`)) return;
    try { await deleteProject(project.id); setNotice("Projeto excluído."); await refresh(); }
    catch (cause) { setNotice(cause instanceof Error ? cause.message : "Falha ao excluir projeto."); }
  };

  const createNewSchedule = async () => {
    if (!scheduleObjective.trim()) return;
    try { await createSchedule({ objective: scheduleObjective, interval_seconds: Number(scheduleInterval), project_id: scheduleProject || undefined, enabled: true }); setScheduleObjective(""); setNotice("Schedule criado; o worker usará o intervalo configurado."); await refresh(); }
    catch (cause) { setNotice(cause instanceof Error ? cause.message : "Falha ao criar schedule."); }
  };

  const removeSchedule = async (schedule: AgentSchedule) => {
    if (!window.confirm("Excluir este schedule?")) return;
    try { await deleteSchedule(schedule.id); setNotice("Schedule excluído."); await refresh(); }
    catch (cause) { setNotice(cause instanceof Error ? cause.message : "Falha ao excluir schedule."); }
  };

  const renderContent = () => {
    if (loading) return <div className="rounded-xl border border-dashed border-neutral-300 px-6 py-12 text-center text-sm text-neutral-500 dark:border-neutral-700">Consultando contratos agentic…</div>;
    if (kind === "projects") return <div className="space-y-3">{projects.length ? projects.map((project) => <ResourceRow key={project.id} onDelete={() => void removeProject(project)}><div className="flex flex-wrap items-center justify-between gap-2"><div><p className="font-medium text-neutral-900 dark:text-white">{editingProject === project.id ? <input autoFocus value={editingProjectName} onChange={(event) => setEditingProjectName(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter") void saveProject(project); if (event.key === "Escape") setEditingProject(null); }} className="h-8 rounded-lg border border-neutral-300 bg-transparent px-2 text-sm dark:border-neutral-700" /> : project.name}</p><p className="mt-1 text-xs text-neutral-500">{project.root || "workspace local gerenciado"} · atualizado {new Date(project.updated_at).toLocaleString()}</p></div>{editingProject === project.id ? <button onClick={() => void saveProject(project)} className="rounded-lg bg-neutral-900 px-3 py-2 text-xs text-white dark:bg-white dark:text-neutral-900">Salvar</button> : <button onClick={() => { setEditingProject(project.id); setEditingProjectName(project.name); }} className="rounded-lg border border-neutral-200 px-3 py-2 text-xs dark:border-neutral-700">Editar</button>}</div></ResourceRow>) : <EmptyState message="Crie um projeto para manter instruções, arquivos e contexto entre missões." action="Criar projeto" onAction={() => document.getElementById("new-project-name")?.focus()} />}</div>;
    if (kind === "tasks") return missions.length ? <div className="space-y-3">{missions.map((mission) => <ResourceRow key={mission.id}><Link to="/agentic" className="block"><div className="flex flex-wrap items-center justify-between gap-2"><p className="font-medium text-neutral-900 dark:text-white">{mission.objective}</p><span className="rounded-full bg-neutral-100 px-2 py-1 text-[10px] font-medium dark:bg-neutral-800">{mission.state}</span></div><p className="mt-1 text-xs text-neutral-500">{mission.id} · {mission.plan?.length ?? 0} passos · {mission.artifacts?.length ?? 0} artifacts · {mission.model || "modelo padrão"}</p></Link></ResourceRow>)}</div> : <EmptyState message="As missões criadas no Agentic Console aparecerão aqui com timeline e artifacts." action="Nova tarefa" onAction={() => window.location.assign("/agentic")} />;
    if (kind === "library") return artifacts.length ? <div className="grid gap-3 md:grid-cols-2">{artifacts.map(({ mission, ...artifact }) => <ResourceRow key={`${mission.id}-${artifact.id}`}><a href={`${API_BASE}/api/agent/v1/missions/${encodeURIComponent(mission.id)}/artifacts/${encodeURIComponent(artifact.id)}`} className="block"><p className="font-medium text-neutral-900 dark:text-white">{artifact.name}</p><p className="mt-1 text-xs text-neutral-500">{Math.round(artifact.size / 1024)} KiB · SHA-256 {artifact.sha256.slice(0, 16)}…</p><p className="mt-1 text-[10px] text-violet-600">Missão {mission.id}</p></a></ResourceRow>)}</div> : <EmptyState message="Os artifacts gerados aparecerão aqui com preview, hash, versão e download." action="Abrir Agentic Console" onAction={() => window.location.assign("/agentic")} />;
    if (kind === "scheduled") return <div className="space-y-3">{schedules.length ? schedules.map((schedule) => <ResourceRow key={schedule.id} onDelete={() => void removeSchedule(schedule)}><div className="flex flex-wrap items-center justify-between gap-2"><div><p className="font-medium text-neutral-900 dark:text-white">{schedule.objective}</p><p className="mt-1 text-xs text-neutral-500">a cada {schedule.interval_seconds}s · próxima {new Date(schedule.next_run_at).toLocaleString()}</p></div><span className={`rounded-full px-2 py-1 text-[10px] ${schedule.enabled ? "bg-emerald-50 text-emerald-700" : "bg-neutral-100 text-neutral-500"}`}>{schedule.enabled ? "ativo" : "pausado"}</span></div></ResourceRow>) : <EmptyState message="Nenhuma automação está agendada. Crie uma para ativar o worker persistente local." action="Agendar tarefa" onAction={() => document.getElementById("new-schedule-objective")?.focus()} />}</div>;
    if (kind === "skills") return skills.length ? <div className="space-y-3">{skills.map((skill) => <ResourceRow key={skill.id}><p className="font-medium text-neutral-900 dark:text-white">{skill.id} <span className="text-xs text-neutral-400">v{skill.version}</span></p><p className="mt-1 text-xs text-neutral-500">{skill.description}</p><p className="mt-1 text-[10px] text-neutral-400">{skill.tools?.length ?? 0} tools · {skill.trusted ? "trusted" : "requer revisão"}</p></ResourceRow>)}</div> : <EmptyState message="Nenhuma skill foi carregada pelo runtime. Instale um manifesto revisado no diretório de skills do servidor." action="Abrir configuração" onAction={() => window.location.assign("/settings#agentic")} />;
    return <div className="space-y-5"><section><h3 className="mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-neutral-500">Connectors allowlisted ({connectors.length})</h3><div className="space-y-3">{connectors.length ? connectors.map((connector) => <ResourceRow key={connector.id}><p className="font-medium text-neutral-900 dark:text-white">{connector.provider} <span className="text-xs text-neutral-400">· {connector.id}</span></p><p className="mt-1 truncate text-xs text-neutral-500">{connector.base_url}</p><p className="mt-1 text-[10px] text-neutral-400">{connector.operations?.length ?? 0} operações · token apenas por ambiente</p></ResourceRow>) : <EmptyState message="Nenhum connector configurado. O registro é server-side e nunca inclui o token." />}</div></section><section><h3 className="mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-neutral-500">MCP local e remoto ({mcpServers.length})</h3><div className="space-y-3">{mcpServers.length ? mcpServers.map((server) => <ResourceRow key={server.id}><p className="font-medium text-neutral-900 dark:text-white">{server.id} <span className="ml-2 rounded-full bg-violet-50 px-2 py-1 text-[10px] text-violet-700 dark:bg-violet-950/30 dark:text-violet-300">{server.transport ?? "stdio"}</span></p><p className="mt-1 truncate font-mono text-xs text-neutral-500">{server.command ?? server.url ?? "endpoint remoto"} {server.command ? (server.args ?? []).join(" ") : ""}</p><p className="mt-1 text-[10px] text-neutral-400">{server.allowed_methods?.length ?? 0} métodos allowlisted · credencial somente no servidor</p></ResourceRow>) : <EmptyState message="Nenhum servidor MCP carregado. O comando ou endpoint precisa estar allowlisted no servidor." />}</div></section><p className="text-xs text-neutral-500">{cliCount} CLI(s) de integração detectado(s). A instalação e o registro continuam protegidos no ambiente do operador.</p></div>;
  };

  return <SidebarLayout title={copy.title} sidebar={<AppSidebar current={current} />}><div className="min-h-0 flex-1 overflow-y-auto bg-neutral-50 dark:bg-neutral-900"><div className="mx-auto w-full max-w-6xl px-6 pb-14 pt-10 lg:px-12"><div className="flex flex-col gap-5 border-b border-neutral-200 pb-8 dark:border-neutral-800 md:flex-row md:items-end md:justify-between"><div className="max-w-2xl"><div className="mb-3 flex items-center gap-2 text-xs font-medium text-violet-600 dark:text-violet-300"><SparklesIcon className="h-4 w-4" />{copy.eyebrow}</div><h2 className="font-rounded text-3xl font-semibold tracking-tight text-neutral-950 dark:text-white">{copy.title}</h2><p className="mt-3 text-sm leading-6 text-neutral-500 dark:text-neutral-400">{copy.description}</p></div><button onClick={handleAction} className="inline-flex items-center justify-center gap-2 rounded-xl bg-neutral-950 px-4 py-2.5 text-sm font-medium text-white shadow-sm dark:bg-white dark:text-neutral-950"><PlusIcon className="h-4 w-4" />{copy.action}</button></div>{notice && <div role="status" aria-live="polite" className="mt-5 rounded-xl border border-violet-200 bg-violet-50 px-4 py-3 text-xs leading-5 text-violet-800 dark:border-violet-900/60 dark:bg-violet-950/20 dark:text-violet-200">{notice}</div>}

{kind === "projects" && <section className="mt-6 grid gap-3 rounded-2xl border border-neutral-200/80 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900 md:grid-cols-[1fr_1fr_auto]"><input id="new-project-name" value={projectName} onChange={(event) => setProjectName(event.target.value)} placeholder="Nome do projeto" className="h-10 rounded-xl border border-neutral-300 bg-transparent px-3 text-sm dark:border-neutral-700" /><input value={projectRoot} onChange={(event) => setProjectRoot(event.target.value)} placeholder="Workspace (opcional)" className="h-10 rounded-xl border border-neutral-300 bg-transparent px-3 text-sm dark:border-neutral-700" /><button onClick={() => void createNewProject()} disabled={!projectName.trim()} className="h-10 rounded-xl bg-neutral-900 px-4 text-sm text-white disabled:opacity-40 dark:bg-white dark:text-neutral-900">Criar</button></section>}
{kind === "scheduled" && <section className="mt-6 grid gap-3 rounded-2xl border border-neutral-200/80 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900 md:grid-cols-[1.4fr_0.5fr_0.8fr_auto]"><input id="new-schedule-objective" value={scheduleObjective} onChange={(event) => setScheduleObjective(event.target.value)} placeholder="Objetivo recorrente" className="h-10 rounded-xl border border-neutral-300 bg-transparent px-3 text-sm dark:border-neutral-700" /><input type="number" min="1" value={scheduleInterval} onChange={(event) => setScheduleInterval(event.target.value)} aria-label="Intervalo em segundos" className="h-10 rounded-xl border border-neutral-300 bg-transparent px-3 text-sm dark:border-neutral-700" /><select value={scheduleProject} onChange={(event) => setScheduleProject(event.target.value)} aria-label="Projeto do schedule" className="h-10 rounded-xl border border-neutral-300 bg-transparent px-3 text-sm dark:border-neutral-700"><option value="">Sem projeto</option>{projects.map((project) => <option key={project.id} value={project.id}>{project.name}</option>)}</select><button onClick={() => void createNewSchedule()} disabled={!scheduleObjective.trim()} className="h-10 rounded-xl bg-neutral-900 px-4 text-sm text-white disabled:opacity-40 dark:bg-white dark:text-neutral-900">Agendar</button></section>}

<div className="mt-7 grid gap-3 sm:grid-cols-3"><StatCard label="Status" value={kind === "tasks" ? `${runningMissions} em execução` : kind === "projects" ? `${projects.length} projetos` : kind === "scheduled" ? `${schedules.length} schedules` : kind === "library" ? `${artifacts.length} artifacts` : "Catálogo real"} tone="green" /><StatCard label="Segurança" value="Approvals ativos" tone="violet" /><StatCard label="Persistência" value="Local-first" /></div><div className="mt-7 grid gap-5 lg:grid-cols-[1.45fr_0.8fr]"><section className="rounded-2xl border border-neutral-200/80 bg-neutral-50/50 p-6 dark:border-neutral-800 dark:bg-neutral-950/30"><div className="flex items-center justify-between"><div><h3 className="text-sm font-semibold text-neutral-900 dark:text-white">Dados persistidos</h3><p className="mt-1 text-xs text-neutral-500 dark:text-neutral-400">Esta superfície consulta o contrato agentic, não fixtures visuais.</p></div><CheckCircleIcon className="h-5 w-5 text-emerald-500" /></div><div className="mt-5">{renderContent()}</div></section><aside className="space-y-5"><section className="rounded-2xl border border-neutral-200/80 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900"><div className="flex items-center gap-2 text-sm font-semibold text-neutral-900 dark:text-white"><LockClosedIcon className="h-4 w-4 text-emerald-500" />Política ativa</div><ul className="mt-4 space-y-3 text-xs leading-5 text-neutral-500 dark:text-neutral-400"><li className="flex gap-2"><CheckCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />Secrets não aparecem na interface.</li><li className="flex gap-2"><CheckCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />Ações externas exigem approval.</li><li className="flex gap-2"><CheckCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />Cross-tenant é rejeitado no servidor.</li></ul></section><section className="rounded-2xl border border-neutral-200/80 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900"><div className="flex items-center gap-2 text-sm font-semibold text-neutral-900 dark:text-white"><ClockIcon className="h-4 w-4 text-violet-500" />Próximos passos</div><p className="mt-3 text-xs leading-5 text-neutral-500 dark:text-neutral-400">Para criar uma missão complexa, use o console com seleção de provider, projeto, timeline e approvals.</p><Link to="/agentic" className="mt-4 inline-flex items-center gap-2 text-xs font-medium text-violet-600 dark:text-violet-300">Abrir Agentic Console <ArrowRightIcon className="h-3.5 w-3.5" /></Link></section></aside></div></div></div></SidebarLayout>;
}
