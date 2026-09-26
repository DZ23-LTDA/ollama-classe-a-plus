import { useState } from "react";
import type { AgentConnectorCatalogEntry } from "@/lib/agenticClient";
import {
  connectWithKey,
  disconnectConnector,
  keyHelpURL,
} from "@/lib/connectorConnect";

// ConnectorQuickConnect is shown inside a catalog card. Services that accept
// an API key connect right here; OAuth-only services say plainly what is
// missing instead of sending the user through other pages.
export function ConnectorQuickConnect({
  entry,
  connected,
  onChanged,
}: {
  entry: AgentConnectorCatalogEntry;
  connected: boolean;
  onChanged: () => void;
}) {
  const [key, setKey] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const run = async (action: () => Promise<void>) => {
    setBusy(true);
    setError(null);
    try {
      await action();
      setKey("");
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  if (!entry.api_base_url) {
    return (
      <div className="mt-4 rounded-xl bg-neutral-50 p-3 text-xs leading-5 text-neutral-600 dark:bg-neutral-900 dark:text-neutral-300">
        <p>
          {entry.name} usa login OAuth, que precisa de um app registrado no{" "}
          {entry.name} pela equipe DZ23. Esse login de um clique ainda não está
          disponível para este serviço.
        </p>
      </div>
    );
  }

  const help = keyHelpURL(entry.id);
  return (
    <div className="mt-4 rounded-xl bg-neutral-50 p-3 text-xs leading-5 text-neutral-600 dark:bg-neutral-900 dark:text-neutral-300">
      {connected ? (
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span>
            Conectado com chave de API. O agente pede aprovação antes de ações
            que alteram dados.
          </span>
          <button
            type="button"
            disabled={busy}
            onClick={() => void run(() => disconnectConnector(entry.id))}
            className="rounded-lg px-3 py-1.5 text-red-600 hover:bg-red-50 disabled:opacity-50 dark:text-red-400 dark:hover:bg-red-950/30"
          >
            Desconectar
          </button>
        </div>
      ) : (
        <form
          className="flex flex-wrap items-center gap-2"
          onSubmit={(event) => {
            event.preventDefault();
            void run(() => connectWithKey(entry.id, key));
          }}
        >
          <input
            type="password"
            autoComplete="off"
            spellCheck={false}
            value={key}
            onChange={(event) => setKey(event.target.value)}
            placeholder={`Chave de API do ${entry.name}`}
            aria-label={`Chave de API do ${entry.name}`}
            className="min-w-0 flex-1 rounded-lg border border-neutral-300 bg-white px-3 py-1.5 text-xs text-neutral-900 outline-none dark:border-neutral-700 dark:bg-neutral-950 dark:text-neutral-100"
          />
          <button
            type="submit"
            disabled={busy || !key.trim()}
            className="rounded-lg bg-neutral-900 px-3 py-1.5 text-xs text-white disabled:opacity-40 dark:bg-white dark:text-neutral-900"
          >
            {busy ? "Conectando…" : "Conectar"}
          </button>
          {help && (
            <a
              href={help}
              target="_blank"
              rel="noreferrer"
              className="w-full text-violet-600 hover:underline dark:text-violet-300"
            >
              Onde pego a chave do {entry.name}?
            </a>
          )}
        </form>
      )}
      {error && (
        <p role="alert" className="mt-2 text-red-600 dark:text-red-400">
          {error}
        </p>
      )}
    </div>
  );
}
