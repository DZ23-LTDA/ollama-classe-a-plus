import { useState } from "react";
import { decideCompanyApproval } from "@/lib/agenticClient";
import type { AgentCompany, CompanyApproval } from "@/lib/agenticClient";

export function CompanyApprovalQueue({ company, onCompanyChange }: { company: AgentCompany; onCompanyChange: (company: AgentCompany) => void }) {
  const [busyID, setBusyID] = useState("");
  const [notice, setNotice] = useState("");
  const pending = (company.approvals ?? []).filter((approval) => approval.status === "pending" && approval.resource_type === "spend");
  const decide = async (approval: CompanyApproval, approved: boolean) => {
    setBusyID(approval.id);
    setNotice("");
    try {
      const updated = await decideCompanyApproval(company.id, approval.id, approved, approval.nonce);
      onCompanyChange(updated);
      setNotice(approved ? "Gasto aprovado e contabilizado." : "Gasto rejeitado.");
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "Não foi possível decidir o gasto.");
    } finally {
      setBusyID("");
    }
  };
  return <section className="mt-5 rounded-2xl border border-amber-200/80 bg-amber-50/50 p-5 dark:border-amber-900/60 dark:bg-amber-950/10">
    <div className="flex items-center justify-between gap-3"><div><div className="text-xs font-medium uppercase tracking-[0.14em] text-amber-700 dark:text-amber-300">Approval queue</div><h2 className="mt-1 text-lg font-semibold text-neutral-950 dark:text-white">Gastos aguardando decisão</h2></div><span className="rounded-full bg-amber-100 px-2 py-1 text-xs text-amber-800 dark:bg-amber-900/40 dark:text-amber-200">{pending.length}</span></div>
    {notice && <p role="status" aria-live="polite" className="mt-3 rounded-lg bg-white px-3 py-2 text-xs text-amber-800 dark:bg-neutral-900 dark:text-amber-200">{notice}</p>}
    {pending.length === 0 ? <p className="mt-3 text-xs text-neutral-500">Nenhum gasto pendente. O valor só entra no budget após decisão válida.</p> : <div className="mt-4 space-y-2">{pending.map((approval) => <div key={approval.id} className="rounded-xl border border-amber-200 bg-white p-3 dark:border-amber-900/60 dark:bg-neutral-900"><div className="flex flex-wrap justify-between gap-2 text-xs"><span>{approval.category || approval.policy}</span><span>{approval.amount_cents ?? 0} cents</span></div><p className="mt-1 text-[11px] text-neutral-500">Expira: {approval.expires_at ? new Date(approval.expires_at).toLocaleString() : "não informado"}</p><div className="mt-3 flex gap-2"><button type="button" disabled={busyID !== ""} onClick={() => void decide(approval, true)} className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-medium text-white disabled:opacity-40">Aprovar</button><button type="button" disabled={busyID !== ""} onClick={() => void decide(approval, false)} className="rounded-lg border border-red-300 px-3 py-1.5 text-xs text-red-700 disabled:opacity-40 dark:border-red-800 dark:text-red-300">Rejeitar</button></div></div>)}</div>}
  </section>;
}
