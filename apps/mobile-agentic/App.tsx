import AsyncStorage from "@react-native-async-storage/async-storage";
import * as SecureStore from "expo-secure-store";
import { StatusBar } from "expo-status-bar";
import { useEffect, useMemo, useState } from "react";
import { ActivityIndicator, Pressable, SafeAreaView, ScrollView, StyleSheet, Text, TextInput, View } from "react-native";

type Mission = { id: string; objective: string; state: string; approvals?: Array<{ id: string; step_id: string; status: string }>; last_error?: string };
type Event = { id: string; type: string; step_id?: string; created_at: string };
type Session = { access_token: string; user: { email: string }; organization: { name: string } };
type QueuedAction = { id: string; path: string; method: string; body?: string; token?: string; created_at: string };
const queueKey = "dz23.agent.offline.queue";
const missionKey = "dz23.agent.cached.mission";

async function request<T>(base: string, path: string, token?: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  if (init?.headers && !(init.headers instanceof Headers)) Object.assign(headers, init.headers);
  const response = await fetch(`${base.replace(/\/$/, "")}${path}`, { ...init, headers });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error ?? response.statusText);
  return body as T;
}

export default function App() {
  const [base, setBase] = useState("http://localhost:11434");
  const [draftBase, setDraftBase] = useState(base);
  const [token, setToken] = useState("");
  const [draftToken, setDraftToken] = useState("");
  const [objective, setObjective] = useState("");
  const [mission, setMission] = useState<Mission | null>(null);
  const [events, setEvents] = useState<Event[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [online, setOnline] = useState(true);
  const [queued, setQueued] = useState(0);

  const loadQueue = async () => { const raw = await AsyncStorage.getItem(queueKey); const items: QueuedAction[] = raw ? JSON.parse(raw) : []; setQueued(items.length); return items; };
  const enqueue = async (path: string, method: string, body?: unknown) => { const items = await loadQueue(); items.push({ id: `offline_${Date.now()}_${Math.random().toString(36).slice(2)}`, path, method, body: body ? JSON.stringify(body) : undefined, token, created_at: new Date().toISOString() }); await AsyncStorage.setItem(queueKey, JSON.stringify(items)); setQueued(items.length); };
  const flushQueue = async () => { const items = await loadQueue(); const remaining: QueuedAction[] = []; for (const item of items) { try { await request(base, item.path, item.token, { method: item.method, body: item.body }); } catch { remaining.push(item); } } await AsyncStorage.setItem(queueKey, JSON.stringify(remaining)); setQueued(remaining.length); if (remaining.length === 0) setOnline(true); };

  const refresh = async (missionId = mission?.id) => {
    if (!missionId) return;
    try {
      const [nextMission, nextEvents] = await Promise.all([request<Mission>(base, `/api/agent/v1/missions/${encodeURIComponent(missionId)}`, token), request<{ events: Event[] }>(base, `/api/agent/v1/missions/${encodeURIComponent(missionId)}/events`, token)]);
      setMission(nextMission); setEvents(nextEvents.events); setOnline(true); await AsyncStorage.setItem(missionKey, JSON.stringify({ mission: nextMission, events: nextEvents.events })); void flushQueue();
    } catch { setOnline(false); const cached = await AsyncStorage.getItem(missionKey); if (cached && !mission) { const value = JSON.parse(cached) as { mission: Mission; events: Event[] }; setMission(value.mission); setEvents(value.events); } }
  };

  useEffect(() => { void AsyncStorage.getItem("dz23.agent.base").then((value) => { if (value) { setBase(value); setDraftBase(value); } }); void SecureStore.getItemAsync("dz23.agent.token").then((value) => { if (value) { setToken(value); setDraftToken(value); } }); void loadQueue(); void AsyncStorage.getItem(missionKey).then((raw) => { if (raw) { const value = JSON.parse(raw) as { mission: Mission; events: Event[] }; setMission(value.mission); setEvents(value.events); } }); }, []);
  useEffect(() => { const timer = setInterval(() => void refresh().catch(() => undefined), 3000); return () => clearInterval(timer); }, [mission?.id, base, token]);

  const approvals = useMemo(() => mission?.approvals?.filter((approval) => approval.status === "PENDING") ?? [], [mission]);
  const perform = async (path: string, method: string, body: unknown, fallback: string) => { try { await request(base, path, token, { method, body: JSON.stringify(body) }); setOnline(true); await flushQueue(); return true; } catch { await enqueue(path, method, body); setOnline(false); setError(`${fallback}. A ação foi salva e será sincronizada quando houver conexão.`); return false; } };
  const create = async () => { if (!objective.trim()) return; setBusy(true); setError(""); const payload = { objective, auto_run: false }; try { const created = await request<Mission>(base, "/api/agent/v1/missions", token, { method: "POST", body: JSON.stringify(payload) }); setMission(created); setOnline(true); await refresh(created.id); } catch { await enqueue("/api/agent/v1/missions", "POST", payload); setOnline(false); setError("Servidor indisponível. A missão foi salva e será sincronizada quando houver conexão."); } setObjective(""); setBusy(false); };
  const decide = async (approvalId: string, approved: boolean) => { if (!mission) return; setBusy(true); setError(""); await perform(`/api/agent/v1/missions/${mission.id}/approvals/${approvalId}`, "POST", { approved, reason: "Mobile operator" }, "Falha ao decidir approval"); await refresh(); setBusy(false); };
  const run = async () => { if (!mission) return; setBusy(true); setError(""); await perform(`/api/agent/v1/missions/${mission.id}/run`, "POST", {}, "Falha ao executar"); await refresh(); setBusy(false); };
  const saveSession = async () => { const nextBase = draftBase.trim(); if (!nextBase) return; await AsyncStorage.setItem("dz23.agent.base", nextBase); if (draftToken.trim()) await SecureStore.setItemAsync("dz23.agent.token", draftToken.trim()); setBase(nextBase); setToken(draftToken.trim()); await flushQueue(); };
  const clearSession = async () => { await SecureStore.deleteItemAsync("dz23.agent.token"); setToken(""); setDraftToken(""); };

  return <SafeAreaView style={styles.safe}><StatusBar style="auto" /><ScrollView contentContainerStyle={styles.container}>
    <Text style={styles.eyebrow}>DZ23 AGENTIC</Text><Text style={styles.title}>Mission mobile</Text><Text style={styles.subtitle}>Acompanhe, aprove e execute missões do seu dispositivo, mesmo durante interrupções temporárias de rede.</Text>
    <View style={styles.sync}><View style={[styles.syncDot, { backgroundColor: online ? "#059669" : "#d97706" }]} /><Text style={styles.syncText}>{online ? "Online" : "Offline — cache local ativo"}{queued ? ` · ${queued} ação(ões) pendente(s)` : ""}</Text></View>
    <View style={styles.card}><Text style={styles.label}>Servidor</Text><TextInput value={draftBase} onChangeText={setDraftBase} autoCapitalize="none" autoCorrect={false} style={styles.input} /><Text style={styles.label}>Token Bearer (armazenado no SecureStore)</Text><TextInput value={draftToken} onChangeText={setDraftToken} autoCapitalize="none" autoCorrect={false} secureTextEntry style={styles.input} /><View style={styles.row}><Pressable onPress={() => void saveSession()} style={[styles.secondary, { flex: 1 }]}><Text style={styles.secondaryText}>Salvar sessão</Text></Pressable><Pressable onPress={() => void clearSession()} style={styles.secondary}><Text style={styles.secondaryText}>Sair</Text></Pressable></View></View>
    <View style={styles.card}><Text style={styles.label}>Novo objetivo</Text><TextInput value={objective} onChangeText={setObjective} multiline placeholder="Ex.: verificar os testes do projeto" style={[styles.input, styles.multiline]} /><Pressable disabled={busy || !objective.trim()} onPress={() => void create()} style={[styles.primary, (!objective.trim() || busy) && styles.disabled]}>{busy ? <ActivityIndicator color="#fff" /> : <Text style={styles.primaryText}>Criar missão</Text>}</Pressable></View>
    {error ? <Text style={styles.error}>{error}</Text> : null}
    {mission ? <View style={styles.card}><View style={styles.row}><View style={{ flex: 1 }}><Text style={styles.muted}>{mission.id}</Text><Text style={styles.mission}>{mission.objective}</Text></View><Text style={styles.status}>{mission.state}</Text></View><Pressable onPress={() => void run()} disabled={busy || approvals.length > 0 || mission.state === "COMPLETED"} style={[styles.secondary, (busy || approvals.length > 0) && styles.disabled]}><Text style={styles.secondaryText}>Executar missão</Text></Pressable>{approvals.map((approval) => <View key={approval.id} style={styles.approval}><Text style={styles.label}>Approval: {approval.step_id}</Text><View style={styles.row}><Pressable onPress={() => void decide(approval.id, true)} style={styles.approve}><Text style={styles.primaryText}>Aprovar</Text></Pressable><Pressable onPress={() => void decide(approval.id, false)} style={styles.reject}><Text style={styles.primaryText}>Rejeitar</Text></Pressable></View></View>)}<Text style={styles.label}>Timeline</Text>{events.map((event) => <View key={event.id} style={styles.event}><View style={styles.dot} /><View><Text style={styles.eventType}>{event.type}</Text><Text style={styles.muted}>{event.step_id ?? "mission"} · {new Date(event.created_at).toLocaleString()}</Text></View></View>)}</View> : null}
  </ScrollView></SafeAreaView>;
}

const styles = StyleSheet.create({ safe: { flex: 1, backgroundColor: "#f7f7f5" }, container: { padding: 20, gap: 16 }, eyebrow: { color: "#737373", fontSize: 12, letterSpacing: 2, fontWeight: "700" }, title: { color: "#171717", fontSize: 30, fontWeight: "700" }, subtitle: { color: "#525252", fontSize: 15, lineHeight: 22 }, sync: { flexDirection: "row", alignItems: "center", gap: 8 }, syncDot: { width: 9, height: 9, borderRadius: 5 }, syncText: { color: "#525252", fontSize: 13 }, card: { backgroundColor: "#fff", borderRadius: 18, padding: 16, gap: 12, shadowColor: "#000", shadowOpacity: 0.06, shadowRadius: 12, elevation: 2 }, label: { color: "#404040", fontSize: 13, fontWeight: "600" }, muted: { color: "#737373", fontSize: 12 }, input: { borderColor: "#d4d4d4", borderWidth: 1, borderRadius: 12, padding: 12, color: "#171717", fontSize: 15 }, multiline: { minHeight: 90, textAlignVertical: "top" }, primary: { backgroundColor: "#171717", borderRadius: 12, minHeight: 44, alignItems: "center", justifyContent: "center", paddingHorizontal: 16 }, primaryText: { color: "#fff", fontSize: 14, fontWeight: "700" }, secondary: { borderColor: "#d4d4d4", borderWidth: 1, borderRadius: 12, minHeight: 42, alignItems: "center", justifyContent: "center", paddingHorizontal: 14 }, secondaryText: { color: "#262626", fontSize: 14, fontWeight: "600" }, disabled: { opacity: 0.4 }, error: { color: "#b91c1c", backgroundColor: "#fee2e2", borderRadius: 12, padding: 12 }, row: { flexDirection: "row", alignItems: "center", gap: 10 }, mission: { color: "#171717", fontSize: 16, fontWeight: "600", marginTop: 4 }, status: { color: "#525252", backgroundColor: "#f5f5f5", borderRadius: 20, paddingVertical: 6, paddingHorizontal: 10, fontSize: 12 }, approval: { backgroundColor: "#fffbeb", borderColor: "#fde68a", borderWidth: 1, borderRadius: 12, padding: 12, gap: 10 }, approve: { backgroundColor: "#059669", borderRadius: 10, paddingVertical: 10, paddingHorizontal: 14 }, reject: { backgroundColor: "#dc2626", borderRadius: 10, paddingVertical: 10, paddingHorizontal: 14 }, event: { flexDirection: "row", gap: 10, alignItems: "flex-start", paddingVertical: 8 }, dot: { width: 8, height: 8, borderRadius: 4, backgroundColor: "#737373", marginTop: 5 }, eventType: { color: "#262626", fontSize: 14, fontWeight: "600" } });
