import { describe, expect, it } from "vitest";
import { connectorIcon } from "./connectorIcons";

describe("connectorIcon", () => {
  it("returns brand paths with brand color", () => {
    const icon = connectorIcon("pinterest");
    expect(icon?.path.length).toBeGreaterThan(20);
    expect(icon?.color).toMatch(/^#[0-9A-F]{6}$/i);
  });

  it("keeps near-black brands visible in dark mode", () => {
    expect(connectorIcon("github")?.color).toBe("currentColor");
  });

  it("falls back to null for brands without a free logo", () => {
    expect(connectorIcon("slack")).toBeNull();
    expect(connectorIcon("unknown")).toBeNull();
  });
});
