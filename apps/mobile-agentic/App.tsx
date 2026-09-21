import AsyncStorage from "@react-native-async-storage/async-storage";
import { StatusBar } from "expo-status-bar";
import { useEffect, useMemo, useState } from "react";
import { ActivityIndicator, Pressable, SafeAreaView, ScrollView, StyleSheet, Text, TextInput, View } from "react-native";

type Mission = { id: string; objective: string; state: string; approvals?: Array<{ id: string; step_id: string; status: string }>; last_error?: string };
type Event = { id: string; type: string; step_id?: string; created_at: string };

async function request<T>(base: string, path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${base.replace(/\/$/, "")}${path}`, { ...init, headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) } });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error ?? response.statusText);
  return body as T;
}

export default function App() {
  const [base, setBase] = useState("http://localhost:11434");
  const [draftBase, setDraftBase] = useState(base);
  const [objective, setObjective] = useState("");
  const [mission, setMission] = useState<Mission | null>(null);
  const [events, setEvents] = useState<Event[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const refresh = async (missionId = mission?.id) => {
    if (!missionId) return;
    const [nextMission, nextEvents] = await Promise.all([
      request<Mission>(base, `/api/agent/v1/missions/${encodeURIComponent(missionId)}`),
      request<{ events: Event[] }>(base, `/api/agent/v1/missions/${encodeURIComponent(missionId)}/events`),
    ]);
    setMission(nextMission);
    setEvents(nextEvents.events);
  };

  useEffect(() => { void AsyncStorage.getItem("dz23.agent.base").then((value) => { if (value) { setBase(value); setDraftBase(value); } }); }, []);
  useEffect(() => { const timer = setInterval(() => void refresh().catch(() => undefined), 3000); return () => clearInterval(timer); }, [mission?.id, base]);

  const approvals = useMemo(() => mission?.approvals?.filter((approval) => approval.status === "PENDING") ?? [], [mission]);
  const create = async () => {
    setBusy(true); setError("");
    try {
      const created = await request<Mission>(base, "/api/agent/v1/missions", { method: "POST", body: JSON.stringify({ objective, auto_run: false }) });
      setMission(created); setObjective(""); await refresh(created.id);
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Falha de rede"); } finally { setBusy(false); }
  };
  const decide = async (approvalId: string, approved: boolean) => {
    if (!mission) return;
    setBusy(true); setError("");
    try { await request(base, `/api/agent/v1/missions/${mission.id}/approvals/${approvalId}`, { method: "POST", body: JSON.stringify({ approved, reason: "Mobile operator" }) }); await refresh(); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Falha ao decidir approval"); } finally { setBusy(false); }
  };
  const run = async () => {
    if (!mission) return;
    setBusy(true); setError("");
    try { await request(base, `/api/agent/v1/missions/${mission.id}/run`, { method: "POST", body: "{}" }); await refresh(); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Falha ao executar"); } finally { setBusy(false); }
  };
  const saveBase = async () => { const value = draftBase.trim(); if (!value) return; await AsyncStorage.setItem("dz23.agent.base", value); setBase(value); };

  return <SafeAreaView style={styles.safe}><StatusBar style="auto" /><ScrollView contentContainerStyle={styles.container}>
    <Text style={styles.eyebrow}>DZ23 AGENTIC</Text><Text style={styles.title}>Mission mobile</Text><Text style={styles.subtitle}>Acompanhe, aprove e execute missões do seu dispositivo.</Text>
    <View style={styles.card}><Text style={styles.label}>Servidor</Text><TextInput value={draftBase} onChangeText={setDraftBase} autoCapitalize="none" style={styles.input} /><Pressable onPress={() => void saveBase()} style={styles.secondary}><Text style={styles.secondaryText}>Salvar servidor</Text></Pressable></View>
    <View style={styles.card}><Text style={styles.label}>Novo objetivo</Text><TextInput value={objective} onChangeText={setObjective} multiline placeholder="Ex.: verificar os testes do projeto" style={[styles.input, styles.multiline]} /><Pressable disabled={busy || !objective.trim()} onPress={() => void create()} style={[styles.primary, (!objective.trim() || busy) && styles.disabled]}>{busy ? <ActivityIndicator color="#fff" /> : <Text style={styles.primaryText}>Criar missão</Text>}</Pressable></View>
    {error ? <Text style={styles.error}>{error}</Text> : null}
    {mission ? <View style={styles.card}><View style={styles.row}><View style={{ flex: 1 }}><Text style={styles.muted}>{mission.id}</Text><Text style={styles.mission}>{mission.objective}</Text></View><Text style={styles.status}>{mission.state}</Text></View>
      <Pressable onPress={() => void run()} disabled={busy || approvals.length > 0 || mission.state === "COMPLETED"} style={[styles.secondary, (busy || approvals.length > 0) && styles.disabled]}><Text style={styles.secondaryText}>Executar missão</Text></Pressable>
      {approvals.map((approval) => <View key={approval.id} style={styles.approval}><Text style={styles.label}>Approval: {approval.step_id}</Text><View style={styles.row}><Pressable onPress={() => void decide(approval.id, true)} style={styles.approve}><Text style={styles.primaryText}>Aprovar</Text></Pressable><Pressable onPress={() => void decide(approval.id, false)} style={styles.reject}><Text style={styles.primaryText}>Rejeitar</Text></Pressable></View></View>)}
      <Text style={styles.label}>Timeline</Text>{events.map((event) => <View key={event.id} style={styles.event}><View style={styles.dot} /><View><Text style={styles.eventType}>{event.type}</Text><Text style={styles.muted}>{event.step_id ?? "mission"} · {new Date(event.created_at).toLocaleString()}</Text></View></View>)}
    </View> : null}
  </ScrollView></SafeAreaView>;
}

const styles = StyleSheet.create({ safe: { flex: 1, backgroundColor: "#f7f7f5" }, container: { padding: 20, gap: 16 }, eyebrow: { color: "#737373", fontSize: 12, letterSpacing: 2, fontWeight: "700" }, title: { color: "#171717", fontSize: 30, fontWeight: "700" }, subtitle: { color: "#525252", fontSize: 15, lineHeight: 22 }, card: { backgroundColor: "#fff", borderRadius: 18, padding: 16, gap: 12, shadowColor: "#000", shadowOpacity: 0.06, shadowRadius: 12, elevation: 2 }, label: { color: "#404040", fontSize: 13, fontWeight: "600" }, muted: { color: "#737373", fontSize: 12 }, input: { borderColor: "#d4d4d4", borderWidth: 1, borderRadius: 12, padding: 12, color: "#171717", fontSize: 15 }, multiline: { minHeight: 90, textAlignVertical: "top" }, primary: { backgroundColor: "#171717", borderRadius: 12, minHeight: 44, alignItems: "center", justifyContent: "center", paddingHorizontal: 16 }, primaryText: { color: "#fff", fontSize: 14, fontWeight: "700" }, secondary: { borderColor: "#d4d4d4", borderWidth: 1, borderRadius: 12, minHeight: 42, alignItems: "center", justifyContent: "center", paddingHorizontal: 14 }, secondaryText: { color: "#262626", fontSize: 14, fontWeight: "600" }, disabled: { opacity: 0.4 }, error: { color: "#b91c1c", backgroundColor: "#fee2e2", borderRadius: 12, padding: 12 }, row: { flexDirection: "row", alignItems: "center", gap: 10 }, mission: { color: "#171717", fontSize: 16, fontWeight: "600", marginTop: 4 }, status: { color: "#525252", backgroundColor: "#f5f5f5", borderRadius: 20, paddingVertical: 6, paddingHorizontal: 10, fontSize: 12 }, approval: { backgroundColor: "#fffbeb", borderColor: "#fde68a", borderWidth: 1, borderRadius: 12, padding: 12, gap: 10 }, approve: { backgroundColor: "#059669", borderRadius: 10, paddingVertical: 10, paddingHorizontal: 14 }, reject: { backgroundColor: "#dc2626", borderRadius: 10, paddingVertical: 10, paddingHorizontal: 14 }, event: { flexDirection: "row", gap: 10, alignItems: "flex-start", paddingVertical: 8 }, dot: { width: 8, height: 8, borderRadius: 4, backgroundColor: "#737373", marginTop: 5 }, eventType: { color: "#262626", fontSize: 14, fontWeight: "600" }
});
