import { describe, expect, it } from "vitest";
import type { AgentConnectorCatalogEntry } from "./agenticClient";
import { connectorState, filterConnectors } from "./connectors";

const entry = (
  id: string,
  kind: string,
  name = id,
): AgentConnectorCatalogEntry => ({
  id,
  name,
  kind,
  category: "Teste",
  description: `${name} descrição`,
  auth: "oauth",
  source: "built-in",
  status: "available",
});

const catalog = [
  entry("github", "app", "GitHub"),
  entry("jira", "custom_api", "Jira"),
  entry("composio", "mcp", "Composio"),
];

describe("connectors", () => {
  it("derives state from registered connectors and MCP servers", () => {
    expect(connectorState(catalog[0], [], [])).toBe("available");
    expect(
      connectorState(
        catalog[0],
        [{ id: "gh", provider: "github", base_url: "https://api.github.com" }],
        [],
      ),
    ).toBe("connected");
    expect(
      connectorState(
        catalog[0],
        [{ id: "github", provider: "github", base_url: "x", disabled: true }],
        [],
      ),
    ).toBe("disabled");
    expect(connectorState(catalog[2], [], [{ id: "composio" }])).toBe(
      "connected",
    );
  });

  it("filters by tab and search text", () => {
    expect(
      filterConnectors(catalog, "app", "", [], []).map((e) => e.id),
    ).toEqual(["github"]);
    expect(
      filterConnectors(catalog, "custom_api", "JIRA", [], []).map((e) => e.id),
    ).toEqual(["jira"]);
    expect(filterConnectors(catalog, "app", "jira", [], [])).toEqual([]);
    expect(
      filterConnectors(catalog, "connected", "", [], [{ id: "composio" }]).map(
        (e) => e.id,
      ),
    ).toEqual(["composio"]);
  });
});
