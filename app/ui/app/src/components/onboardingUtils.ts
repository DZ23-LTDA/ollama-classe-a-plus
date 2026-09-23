import type { ClaudeDesktopStatus } from "@/types/webview";

export const FIRST_MODEL_COMMAND = "ollama";

export function shouldShowClaudeConnectedIntro(status: ClaudeDesktopStatus) {
  return status.connected && !status.startFailed && !status.used;
}
