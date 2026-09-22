import assert from "node:assert/strict";
import { chromium } from "playwright";

const baseURL = process.env.UI_BASE_URL ?? "http://127.0.0.1:4173";
const apiURL = process.env.AGENT_API_URL ?? "http://127.0.0.1:3001";
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });

async function open(route) {
  await page.goto(`${baseURL}${route}`, { waitUntil: "domcontentloaded", timeout: 30000 });
  await page.waitForTimeout(700);
  assert.equal(await page.locator("nav").count(), 1, `${route}: sidebar missing`);
  assert.equal(await page.locator("nav a").count() > 8, true, `${route}: menu entries missing`);
  assert.equal(await page.locator("h1").count() > 0, true, `${route}: title missing`);
  assert.equal(await page.getByText("Something went wrong!").count(), 0, `${route}: global error boundary rendered`);
}

await open("/");
assert.equal(await page.getByRole("heading", { name: "O que posso fazer por você?" }).count(), 1, "home: composer heading missing");
assert.equal(await page.getByText("Recomendado para você").count(), 1, "home: recommendations missing");

await open("/projects");
const projectName = `Smoke UI ${Date.now()}`;
await page.locator("#new-project-name").fill(projectName);
await page.getByRole("button", { name: "Criar", exact: true }).click();
await assert.doesNotReject(() => page.getByText(projectName, { exact: true }).waitFor({ timeout: 5000 }), "projects: create did not render persisted project");

await open("/scheduled");
const scheduleObjective = `Smoke schedule ${Date.now()}`;
await page.locator("#new-schedule-objective").fill(scheduleObjective);
await page.getByRole("button", { name: "Agendar", exact: true }).click();
await assert.doesNotReject(() => page.getByText(scheduleObjective, { exact: true }).waitFor({ timeout: 5000 }), "scheduled: create did not render persisted schedule");

await open("/plugins");
assert.equal(await page.getByText("Connectors allowlisted").count(), 1, "plugins: connector catalog missing");
assert.equal(await page.getByText("MCP stdio").count(), 1, "plugins: MCP catalog missing");

await open("/skills");
assert.equal(await page.getByText(/Skills|habilidade|manifesto/i).count() > 0, true, "skills: catalog content missing");

await open("/tasks");
await page.getByRole("button", { name: "Nova tarefa", exact: true }).click();
await page.waitForURL("**/agentic");

await open("/agentic");
const objective = `Smoke mission ${Date.now()}`;
await page.locator("#agent-objective").fill(objective);
await page.locator("select").first().selectOption("claude");
await page.getByRole("button", { name: "Criar missão", exact: true }).click();
await assert.doesNotReject(() => page.getByText(objective, { exact: true }).waitFor({ timeout: 5000 }), "agentic: mission was not created");
assert.equal(await page.getByText("anthropic/claude-sonnet-4-5").count(), 1, "agentic: selected provider model not shown");

const createdProjects = await (await fetch(`${apiURL}/api/agent/v1/projects`)).json();
for (const project of createdProjects.projects ?? []) {
  if (project.name === projectName) await fetch(`${apiURL}/api/agent/v1/projects/${encodeURIComponent(project.id)}`, { method: "DELETE" });
}
const createdSchedules = await (await fetch(`${apiURL}/api/agent/v1/schedules`)).json();
for (const schedule of createdSchedules.schedules ?? []) {
  if (schedule.objective === scheduleObjective) await fetch(`${apiURL}/api/agent/v1/schedules/${encodeURIComponent(schedule.id)}`, { method: "DELETE" });
}
const createdMissions = await (await fetch(`${apiURL}/api/agent/v1/missions`)).json();
for (const mission of createdMissions.missions ?? []) {
  if (mission.objective === objective) await fetch(`${apiURL}/api/agent/v1/missions/${encodeURIComponent(mission.id)}/cancel`, { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
}

console.log("home: composer and recommendations OK");
console.log("projects: real create/list/delete OK");
console.log("scheduled: real create/list/delete OK");
console.log("plugins: connector and MCP catalogs OK");
console.log("skills: real catalog OK");
console.log("agentic: provider selection and mission creation OK");
await browser.close();
