import { afterEach, describe, expect, it, vi } from "vitest";
import { agentFetch, clearAgentSession, hasAgentSession, setAgentSession } from "./agenticClient";

describe("agent session security", () => {
  afterEach(() => {
    clearAgentSession();
    vi.unstubAllGlobals();
  });

  it("keeps the bearer only in memory and sends organization scope", async () => {
    const dispatchEvent = vi.fn();
    vi.stubGlobal("window", { dispatchEvent });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true }), { status: 200 })));
    setAgentSession("  session-token  ", " org-a ");

    await agentFetch("/api/agent/v1/health");

    const request = vi.mocked(fetch).mock.calls[0]?.[1] as RequestInit;
    expect((request.headers as Record<string, string>).Authorization).toBe("Bearer session-token");
    expect((request.headers as Record<string, string>)["X-Ollama-Organization"]).toBe("org-a");
    expect(hasAgentSession()).toBe(true);
    expect((globalThis.window as unknown as { localStorage?: unknown }).localStorage).toBeUndefined();
  });

  it("clears the session and emits an event on unauthorized responses", async () => {
    const dispatchEvent = vi.fn();
    vi.stubGlobal("window", { dispatchEvent });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: "expired" }), { status: 401 })));
    setAgentSession("session-token", "org-a");

    await expect(agentFetch("/api/agent/v1/projects")).rejects.toThrow("expired");

    expect(hasAgentSession()).toBe(false);
    expect(dispatchEvent).toHaveBeenCalled();
  });

  it("keeps the session on forbidden responses", async () => {
    const dispatchEvent = vi.fn();
    vi.stubGlobal("window", { dispatchEvent });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: "organization denied" }), { status: 403 })));
    setAgentSession("session-token", "org-a");

    await expect(agentFetch("/api/agent/v1/projects/org-b")).rejects.toThrow("organization denied");

    expect(hasAgentSession()).toBe(true);
    expect(dispatchEvent).not.toHaveBeenCalled();
  });

  it("rejects empty session tokens", () => {
    expect(() => setAgentSession("   ")).toThrow("session token is required");
    expect(hasAgentSession()).toBe(false);
  });
});
