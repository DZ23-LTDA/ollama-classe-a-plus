import assert from "node:assert/strict";
import { chromium } from "playwright";

const baseURL = process.env.UI_BASE_URL ?? "http://127.0.0.1:4173";
const routes = ["/projects", "/library", "/scheduled", "/skills", "/plugins", "/tasks"];
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });

for (const route of routes) {
  await page.goto(`${baseURL}${route}`, { waitUntil: "networkidle" });
  assert.equal(await page.locator("nav").count(), 1, `${route}: sidebar missing`);
  assert.equal(await page.locator("nav a").count() > 8, true, `${route}: menu entries missing`);
  assert.equal(await page.locator("h1").count() > 0, true, `${route}: title missing`);

  const actionButton = page.getByRole("button", { name: /Criar|Novo|Agendar|Adicionar|Nova tarefa/ }).first();
  assert.equal(await actionButton.count(), 1, `${route}: primary action missing`);
  await actionButton.click();
  if (route === "/tasks") {
    await page.waitForURL("**/agentic");
  } else {
    assert.equal(await page.getByRole("status").count(), 1, `${route}: primary action did not report its state`);
  }

  await page.goto(`${baseURL}${route}`, { waitUntil: "networkidle" });
  await page.getByRole("button", { name: "Explorar fluxo" }).click();
  if (route === "/projects") {
    await page.waitForURL("**/agentic");
  } else {
    assert.equal(await page.getByRole("status").count(), 1, `${route}: explore action did not report its state`);
  }

  console.log(`${route}: sidebar, title, primary action and empty-state action OK`);
}

await browser.close();
