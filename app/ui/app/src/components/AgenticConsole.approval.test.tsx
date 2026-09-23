import { act, create, type ReactTestInstance } from "react-test-renderer";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mocks = vi.hoisted(() => ({
  agentFetch: vi.fn(),
  getModels: vi.fn(),
  listProjects: vi.fn(),
}));

vi.mock("@/lib/agenticClient", () => ({
  agentFetch: mocks.agentFetch,
  listProjects: mocks.listProjects,
}));

vi.mock("@/api", () => ({
  getModels: mocks.getModels,
}));

import AgenticConsole from "./AgenticConsole";

function textContent(node: ReactTestInstance): string {
  return node.children
    .map((child) => (typeof child === "string" ? child : textContent(child)))
    .join("");
}

const mission = {
  id: "mission-1",
  objective: "Write a release note",
  state: "PLANNED",
  approvals: [
    {
      id: "approval-1",
      step_id: "step-1",
      status: "PENDING",
      nonce: "nonce-1",
    },
  ],
  artifacts: [],
};

describe("AgenticConsole approval decisions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
    vi.stubGlobal("window", {
      setInterval: vi.fn(() => 1),
      clearInterval: vi.fn(),
      location: { search: "" },
    });
    mocks.getModels.mockResolvedValue([]);
    mocks.listProjects.mockResolvedValue({ projects: [] });
    mocks.agentFetch.mockImplementation(async (path: string, init?: RequestInit) => {
      if (path === "/api/agent/v1/metrics") return {};
      if (path === "/api/agent/v1/missions" && init?.method === "POST") return mission;
      if (path === "/api/agent/v1/missions/mission-1") return mission;
      if (path === "/api/agent/v1/missions/mission-1/events") return { events: [] };
      if (path.includes("/approvals/approval-1")) return {};
      throw new Error(`unexpected request: ${path}`);
    });
  });

  it("blocks an empty decision reason and sends the explicit reason", async () => {
    let renderer!: ReturnType<typeof create>;
    await act(async () => {
      renderer = create(<AgenticConsole />);
      await Promise.resolve();
    });

    const objective = renderer.root.findByProps({ id: "agent-objective" });
    await act(async () => {
      objective.props.onChange({ target: { value: "Write a release note" } });
    });

    const createButton = renderer.root
      .findAllByType("button")
      .find((button) => textContent(button).includes("Criar missão"));
    expect(createButton).toBeDefined();
    await act(async () => {
      createButton!.props.onClick();
      await Promise.resolve();
      await Promise.resolve();
    });

    const approveButton = renderer.root
      .findAllByType("button")
      .find((button) => textContent(button).includes("Aprovar"));
    expect(approveButton).toBeDefined();
    expect(approveButton!.props.disabled).toBe(true);

    const reason = renderer.root.findByProps({ id: "approval-reason-approval-1" });
    await act(async () => {
      reason.props.onChange({
        target: { value: "Reviewed the scope and verified the release impact." },
      });
    });

    expect(approveButton!.props.disabled).toBe(false);
    await act(async () => {
      approveButton!.props.onClick();
      await Promise.resolve();
      await Promise.resolve();
    });

    const approvalCall = mocks.agentFetch.mock.calls.find(([path, init]) =>
      String(path).includes("/approvals/approval-1") && init?.method === "POST",
    );
    expect(approvalCall).toBeDefined();
    expect(JSON.parse(approvalCall![1].body)).toEqual({
      approved: true,
      nonce: "nonce-1",
      reason: "Reviewed the scope and verified the release impact.",
    });
  });

  it("announces runtime load failures as an accessible alert", async () => {
    mocks.agentFetch.mockImplementationOnce(async () => {
      throw new Error("runtime unavailable");
    });

    let renderer!: ReturnType<typeof create>;
    await act(async () => {
      renderer = create(<AgenticConsole />);
      await Promise.resolve();
    });

    const alert = renderer.root.findByProps({ role: "alert" });
    expect(alert.props["aria-live"]).toBe("assertive");
    expect(alert.props["aria-atomic"]).toBe("true");
    expect(textContent(alert)).toContain("runtime unavailable");
  });
});
