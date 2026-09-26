import type {
  AgentConnector,
  AgentConnectorCatalogEntry,
  AgentMCPServer,
} from "@/lib/agenticClient";

export type ConnectorTab = "app" | "custom_api" | "mcp" | "connected";

export type ConnectorState = "connected" | "disabled" | "available";

// connectorState reports whether a catalog entry is registered in the runtime.
// Registration is what makes a connector usable; the catalog alone is not.
export function connectorState(
  entry: AgentConnectorCatalogEntry,
  connectors: AgentConnector[],
  mcpServers: AgentMCPServer[],
): ConnectorState {
  const registered =
    connectors.find(
      (item) => item.id === entry.id || item.provider === entry.id,
    ) ?? mcpServers.find((item) => item.id === entry.id);
  if (!registered) return "available";
  return registered.disabled ? "disabled" : "connected";
}

export function filterConnectors(
  catalog: AgentConnectorCatalogEntry[],
  tab: ConnectorTab,
  query: string,
  connectors: AgentConnector[],
  mcpServers: AgentMCPServer[],
): AgentConnectorCatalogEntry[] {
  const needle = query.trim().toLowerCase();
  return catalog.filter((entry) => {
    const inTab =
      tab === "connected"
        ? connectorState(entry, connectors, mcpServers) !== "available"
        : entry.kind === tab;
    if (!inTab) return false;
    if (!needle) return true;
    return [entry.name, entry.description, entry.category]
      .join(" ")
      .toLowerCase()
      .includes(needle);
  });
}
