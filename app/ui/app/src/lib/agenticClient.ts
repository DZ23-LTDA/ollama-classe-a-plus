import { API_BASE } from "@/lib/config";

export type AgentProject = {
  id: string;
  name: string;
  root?: string;
  organization_id?: string;
  created_at: string;
  updated_at: string;
};

export type AgentSchedule = {
  id: string;
  objective: string;
  model?: string;
  workspace?: string;
  project_id?: string;
  organization_id?: string;
  interval_seconds: number;
  enabled: boolean;
  next_run_at: string;
  last_run_at?: string;
};

export type AgentArtifact = {
  id: string;
  name: string;
  sha256: string;
  size: number;
  media_type?: string;
};

export type AgentMission = {
  id: string;
  version: number;
  objective: string;
  model?: string;
  workspace?: string;
  project_id?: string;
  organization_id?: string;
  state: string;
  plan?: Array<{
    id: string;
    title: string;
    kind: string;
    state: string;
    requires_approval: boolean;
  }>;
  approvals?: Array<{ id: string; step_id: string; status: string; reason?: string }>;
  artifacts?: AgentArtifact[];
  last_error?: string;
  created_at: string;
  updated_at: string;
};

export type AgentEvent = {
  id: string;
  type: string;
  step_id?: string;
  created_at: string;
  payload?: unknown;
};

export type AgentConnector = {
  id: string;
  provider: string;
  base_url: string;
  token_env?: string;
  oauth_provider?: string;
  allowed_origins?: string[];
  operations?: Array<{ name: string; methods: string[]; path_prefixes: string[] }>;
};

export type AgentMCPServer = {
  id: string;
  command: string;
  args?: string[];
  allowed_methods?: string[];
  environment_vars?: string[];
  timeout_seconds?: number;
};

export type AgentSkill = {
  id: string;
  version: string;
  description: string;
  scopes?: string[];
  tools?: string[];
  trusted: boolean;
};

export type AgentCLIStatus = {
  id: string;
  name: string;
  section: string;
  visibility: string;
  mode: string;
  executables?: string[];
  installed: boolean;
  executable?: string;
};

function agentHeaders(): Record<string, string> {
  if (typeof window === "undefined") return {};
  const token = window.localStorage.getItem("ollama-agent-token");
  const organization = window.localStorage.getItem("ollama-agent-organization");
  return {
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(organization ? { "X-Ollama-Organization": organization } : {}),
  };
}

export async function agentFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...agentHeaders(),
      ...(init.headers ?? {}),
    },
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    const message = typeof body?.error === "string" ? body.error : response.statusText || "Agent API request failed";
    throw new Error(message);
  }
  return body as T;
}

export const listProjects = () => agentFetch<{ projects: AgentProject[] }>("/api/agent/v1/projects");
export const createProject = (name: string, root = "") =>
  agentFetch<AgentProject>("/api/agent/v1/projects", { method: "POST", body: JSON.stringify({ name, root }) });
export const updateProject = (id: string, name: string, root = "") =>
  agentFetch<AgentProject>(`/api/agent/v1/projects/${encodeURIComponent(id)}`, { method: "PATCH", body: JSON.stringify({ name, root }) });
export const deleteProject = (id: string) =>
  agentFetch<void>(`/api/agent/v1/projects/${encodeURIComponent(id)}`, { method: "DELETE" });

export const listMissions = () => agentFetch<{ missions: AgentMission[] }>("/api/agent/v1/missions");
export const getMission = (id: string) => agentFetch<AgentMission>(`/api/agent/v1/missions/${encodeURIComponent(id)}`);
export const listMissionEvents = (id: string) =>
  agentFetch<{ events: AgentEvent[] }>(`/api/agent/v1/missions/${encodeURIComponent(id)}/events`);
export const createMission = (payload: { objective: string; model?: string; project_id?: string; workspace?: string; auto_run?: boolean }) =>
  agentFetch<AgentMission>("/api/agent/v1/missions", { method: "POST", body: JSON.stringify(payload) });
export const runMission = (id: string) =>
  agentFetch<AgentMission>(`/api/agent/v1/missions/${encodeURIComponent(id)}/run`, { method: "POST", body: "{}" });
export const decideMissionApproval = (missionID: string, approvalID: string, approved: boolean) =>
  agentFetch<AgentMission>(`/api/agent/v1/missions/${encodeURIComponent(missionID)}/approvals/${encodeURIComponent(approvalID)}`, {
    method: "POST",
    body: JSON.stringify({ approved, reason: approved ? "Aprovado no Agentic Console" : "Rejeitado no Agentic Console" }),
  });

export const listSchedules = () => agentFetch<{ schedules: AgentSchedule[] }>("/api/agent/v1/schedules");
export const createSchedule = (payload: Partial<AgentSchedule>) =>
  agentFetch<AgentSchedule>("/api/agent/v1/schedules", { method: "POST", body: JSON.stringify(payload) });
export const updateSchedule = (id: string, payload: Partial<AgentSchedule>) =>
  agentFetch<AgentSchedule>(`/api/agent/v1/schedules/${encodeURIComponent(id)}`, { method: "PATCH", body: JSON.stringify(payload) });
export const deleteSchedule = (id: string) =>
  agentFetch<void>(`/api/agent/v1/schedules/${encodeURIComponent(id)}`, { method: "DELETE" });

export const listSkills = async () => {
  const result = await agentFetch<{ skills?: AgentSkill[] | null }>("/api/agent/v1/skills");
  return { skills: Array.isArray(result.skills) ? result.skills : [] };
};
export const listConnectors = async () => {
  const result = await agentFetch<{ connectors?: AgentConnector[] | null }>("/api/agent/v1/connectors");
  return { connectors: Array.isArray(result.connectors) ? result.connectors : [] };
};
export const listMCPServers = async () => {
  const result = await agentFetch<{ servers?: AgentMCPServer[] | null }>("/api/agent/v1/mcp");
  return { servers: Array.isArray(result.servers) ? result.servers : [] };
};
export const listCLIStatus = async () => {
  const result = await agentFetch<{ tools?: AgentCLIStatus[] | null }>("/api/dz23/cli-catalog");
  return { tools: Array.isArray(result.tools) ? result.tools : [] };
};
